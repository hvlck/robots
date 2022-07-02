package robots

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// IsAllowed checks whether a given *url* is allowed to be scraped for a given RobotList
// the agent corresponds to the User-Agent string
// will return false is there is no valid User-Agent within the RobotList
func (r *RobotList) IsAllowed(url, agent string) bool {
	// matches?
	m := true
	//
	if ent, ok := r.robots[agent]; ok {
		for k, v := range ent.Allowed {
			// matches string w/ wildcards, regardless of whether the URL is allowed/disallowed
			match := true
			// currently iterating over a section that contains a wildcard ("*")
			wildcard := false
			// added to the `url` index to ensure the index of the pattern and url match
			// e.g.
			// 0 1 2 3 4 5 6 7 8 9
			// / * / t e st /
			// 0 1       2 3 4 5 6
			// / t e s t / t e s t
			offset := 0
			// cut is the remaining portion of the url after the current wildcard, used to calculate the offset
			cut := url

			// url matches the pattern exactly
			if url == k {
				m = true
				// exact matches for allowed domains take precedence
				if v {
					break
				}
			} else {
				for i, rn := range []byte(k) {
					idx := i
					// index is only offset if there's a wildcard active
					if wildcard {
						idx += offset
					}

					// check if within index bounds
					if len(url) > idx {
						if rn == '*' {
							wildcard = true
						} else if url[idx] == '/' {
							// end of wildcard matching
							wildcard = false

							// returns remaining portions of url after wildcard portion ("/*/...")
							s := strings.Split(cut, "/")
							if len(s) > 1 {
								cut = strings.Join(s[2:], "/")
								offset = len(s[1]) - 1
							}
						}

						// no wildcard, URLs don't match up
						if url[idx] != rn && !wildcard {
							match = false
							continue
							// characters don't match, wildcard makes it acceptable
						} else if url[idx] != rn && wildcard {
							continue
							// characters match
						} else if url[idx] == rn {
							continue
						}
					}
				}
			}

			// matches string, URL is disallowed
			if match && !v {
				m = false
				// matches string and URL is allowed, or doesn't match and URL isn't allowed
			} else if (match && v) || (!match && !v) {
				m = true
			}
		}

		return m
	} else {
		// no user agent
		return false
	}
}

// Robot is a list of disallowed/allowed URLs for a given user-agent, as well as the crawl delay, given in seconds
type Robot struct {
	Allowed    map[string]bool
	CrawlDelay uint8
}

// RobotList is a list of allowed user agents and sitemaps for a provided `robots.txt` file
type RobotList struct {
	robots   map[string]Robot
	sitemaps []string
}

// parse parses a robots.txt file and generates a list of user agents and sitemaps
func parse(r *bufio.Reader) (RobotList, error) {
	rob := RobotList{
		robots: make(map[string]Robot),
	}

	// initialize list for any site, used as default if no user-agent directive is included in file
	if ent, ok := rob.robots["*"]; !ok {
		ent.Allowed = make(map[string]bool)
		rob.robots["*"] = ent
	}

	// list of user agents that applies to current directives (Allow, Disallow)
	lastUA := []string{"*"}
	for {
		ln, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return RobotList{}, err
		}

		// skip empty lines
		if len(ln) == 1 && ln == "\n" {
			lastUA = []string{""}
			continue
		}

		// trim extra/unnecessary characters
		ln = strings.Trim(strings.Trim(ln, "\n"), " ")

		// ignore comments
		if strings.HasPrefix(ln, "#") {
			continue
		}

		// split directives and values
		f := strings.SplitN(ln, ":", 2)
		directive := f[0]
		// no directive value, ignore line
		if len(f) != 2 {
			continue
		}
		// remove comments at end of line from values
		value := strings.TrimLeft(strings.SplitN(f[1], "#", 2)[0], " ")

		switch strings.ToLower(directive) {
		// url is disallowed
		case "disallow":
			{
				for _, v := range lastUA {
					if ent, ok := rob.robots[v]; ok {
						if len(ent.Allowed) == 0 {
							ent.Allowed = make(map[string]bool)
						}
						ent.Allowed[value] = false
						rob.robots[v] = ent
					}
				}
			}
		// url is allowed
		case "allow":
			{
				for _, v := range lastUA {
					if ent, ok := rob.robots[v]; ok {
						if len(ent.Allowed) == 0 {
							ent.Allowed = make(map[string]bool)
						}
						ent.Allowed[value] = true
						rob.robots[v] = ent
					}
				}
			}
		// sitemap
		case "sitemap":
			{
				rob.sitemaps = append(rob.sitemaps, value)
			}
		// directives apply to user-agent
		case "user-agent":
			{
				// user-agent is present, so default behavior that applies directives to wildcard is removed and
				// the directives are applied to the given user-agent instead
				if len(lastUA) == 1 && lastUA[0] == "*" {
					lastUA[0] = value
				} else {
					lastUA = append(lastUA, value)
				}

				if ent, ok := rob.robots[value]; !ok {
					ent.Allowed = make(map[string]bool)
					rob.robots[value] = ent
				}
			}
		// crawl delay
		case "crawl-delay":
			{
				for _, v := range lastUA {
					if ent, ok := rob.robots[v]; ok {
						n, err := strconv.Atoi(value)
						if err != nil {
							return RobotList{}, err
						}

						ent.CrawlDelay = uint8(n)
						rob.robots[v] = ent
					}
				}
			}
		// invalid directive, move on
		default:
			{
				continue
			}
		}
	}

	return rob, nil
}
