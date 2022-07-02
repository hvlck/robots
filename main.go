package robots

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

func (r *RobotList) IsAllowed(url, agent string) bool {
	m := true
	if ent, ok := r.robots[agent]; ok {
		for k, v := range ent.allowed {
			match := true
			wildcard := false
			offset := 0
			cut := url

			if url == k {
				m = true
				if v {
					break
				}
			} else {
				for i, rn := range []byte(k) {
					idx := i
					if wildcard {
						idx += offset
					}

					if len(url) > idx {
						if rn == '*' {
							wildcard = true
						} else if url[idx] == '/' {
							wildcard = false

							s := strings.Split(cut, "/")
							if len(s) > 1 {
								cut = strings.Join(s[2:], "/")
								offset = len(s[1]) - 1
							}
						}

						if url[idx] != rn && !wildcard {
							match = false
							continue
						} else if url[idx] != rn && wildcard {
							continue
						} else if url[idx] == rn {
							continue
						}
					}
				}
			}

			if match && !v {
				m = false
			} else if (match && v) || (!match && !v) {
				m = true
			}
		}

		return m
	}

	return m
}

type Robot struct {
	allowed     map[string]bool
	crawl_delay uint8
}

type RobotList struct {
	robots   map[string]Robot
	sitemaps []string
}

func parse(r *bufio.Reader) (RobotList, error) {
	rob := RobotList{
		robots: make(map[string]Robot),
	}

	if ent, ok := rob.robots["*"]; !ok {
		ent.allowed = make(map[string]bool)
		rob.robots["*"] = ent
	}

	lastUA := []string{"*"}
	for {
		ln, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return RobotList{}, err
		}

		if len(ln) == 1 && ln == "\n" {
			lastUA = []string{""}
			continue
		}

		ln = strings.Trim(strings.Trim(ln, "\n"), " ")

		if strings.HasPrefix(ln, "#") {
			continue
		}

		f := strings.SplitN(ln, ":", 2)
		directive := f[0]
		// no directive
		if len(f) != 2 {
			continue
		}
		value := strings.TrimLeft(strings.SplitN(f[1], "#", 2)[0], " ")

		switch strings.ToLower(directive) {
		case "disallow":
			{
				for _, v := range lastUA {
					if ent, ok := rob.robots[v]; ok {
						if len(ent.allowed) == 0 {
							ent.allowed = make(map[string]bool)
						}
						ent.allowed[value] = false
						rob.robots[v] = ent
					}
				}
			}
		case "allow":
			{
				for _, v := range lastUA {
					if ent, ok := rob.robots[v]; ok {
						if len(ent.allowed) == 0 {
							ent.allowed = make(map[string]bool)
						}
						ent.allowed[value] = true
						rob.robots[v] = ent
					}
				}
			}
		case "sitemap":
			{
				rob.sitemaps = append(rob.sitemaps, value)
			}
		case "user-agent":
			{
				if len(lastUA) == 1 && lastUA[0] == "*" {
					lastUA[0] = value
				} else {
					lastUA = append(lastUA, value)
				}

				if ent, ok := rob.robots[value]; !ok {
					ent.allowed = make(map[string]bool)
					rob.robots[value] = ent
				}
			}
		case "crawl-delay":
			{
				for _, v := range lastUA {
					if ent, ok := rob.robots[v]; ok {
						n, err := strconv.Atoi(value)
						if err != nil {
							return RobotList{}, err
						}

						ent.crawl_delay = uint8(n)
						rob.robots[v] = ent
					}
				}
			}
		default:
			{
				continue
			}
		}
	}

	return rob, nil
}
