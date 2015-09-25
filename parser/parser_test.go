package parse

import (
	"encoding/json"
	"strings"
	"testing"
)

type ParseTest struct {
	input string
}

var parseTests = map[string]ParseTest{
	"empty": {""},
	"Unnamed": {`
		{
			id,
			name
		}
	`},
	"HeroNameQuery": {`
		query HeroNameQuery {
			hero {
				name
			}
		}
	`},
	"HeroNameAndFriendsQuery": {`
		query HeroNameAndFriendsQuery {
			hero {
				id
				name
				friends {
					name
				}
			}
		}
	`},
	"NestedQuery": {`
		query NestedQuery {
			hero {
				name
				friends {
					name
					appearsIn
					friends {
						name
					}
				}
			}
		}
	`},
	"ArgumentQuery": {`
		query FetchLukeQuery {
			human(id: "1000") {
				name
			}
		}
	`},
	"VariableQuery": {`
		query FetchSomeIDQuery($someId: String!) {
			human(id: $someId) {
				name
			}
		}
	`},
	"AliasedQuery": {`
		query FetchLukeAliased {
			luke: human(id: "1000") {
				name
			}
		}
	`},
	"DoubleQuery": {`
		query FetchLukeAndLeiaAliased {
			luke: human(id: "1000") {
				name
			}
			leia: human(id: "1003") {
				name
			}
		}
     `},
	"FragmentQuery": {`
		query UseFragment {
			luke: human(id: "1000") {
				...HumanFragment
			}
			leia: human(id: "1003") {
				...HumanFragment
			}
		}
		fragment HumanFragment on Human {
			name
			homePlanet
		}
    `},
	"TypenameQuery": {`
		query CheckTypeOfR2 {
			hero {
				__typename
				name
			}
		}
    `},
	"InlineFragments": {`
		query inlineFragmentTyping {
			profiles(handles: ["zuck", "cocacola"]) {
				handle
				... on User {
					friends {
						count
					}
				}
				... on Page {
					likers {
						count
					}
				}
			}
		}
	`},
}

func TestParser(t *testing.T) {
	for name, test := range parseTests {
		actual, err := Reader(strings.NewReader(test.input))
		if err != nil {
			t.Errorf("Error %s: %s", name, err)
		}

		json, _ := json.MarshalIndent(actual, "", "  ")
		t.Logf("%s, %s\n", name, json)
	}
}
