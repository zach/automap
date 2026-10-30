package canonical_test

import (
	"testing"

	"github.com/zach/automap/internal/canonical"
)

// Vectors from RFC 8785 §3.2.3 and appendix B.
func TestCanonicalForm(t *testing.T) {
	for _, c := range []struct{ name, in, want string }{
		{
			"RFC 8785 example",
			`{"numbers":[333333333.33333329,1E30,4.50,2e-3,0.000000000000000000000000001],"string":"\u20ac$\u000F\u000aA'\u0042\u0022\u005c\\\"\/","literals":[null,true,false]}`,
			`{"literals":[null,true,false],"numbers":[333333333.3333333,1e+30,4.5,0.002,1e-27],"string":"€$\u000f\nA'B\"\\\\\"/"}`,
		},
		{
			"keys sort by UTF-16 code units",
			`{"\u20ac":"Euro Sign","\r":"Carriage Return","\ufb33":"Hebrew Letter Dalet With Dagesh","1":"One","\ud83d\ude00":"Emoji: Grinning Face","\u0080":"Control","\u00f6":"Latin Small Letter O With Diaeresis"}`,
			"{\"\\r\":\"Carriage Return\",\"1\":\"One\",\"\u0080\":\"Control\",\"\u00f6\":\"Latin Small Letter O With Diaeresis\",\"\u20ac\":\"Euro Sign\",\"\U0001F600\":\"Emoji: Grinning Face\",\"\ufb33\":\"Hebrew Letter Dalet With Dagesh\"}",
		},
		{"HTML characters are not escaped", `{"a":"<&>","b":" "}`, "{\"a\":\"<&>\",\"b\":\" \"}"},
		{"integers", `[0,-0,1,-12,9007199254740991,1e21,1e-7,123e-2]`, `[0,0,1,-12,9007199254740991,1e+21,1e-7,1.23]`},
		{"nested whitespace", "{ \"b\" : [ 1 , { \"d\":true, \"c\":null } ], \"a\" : \"x\" }", `{"a":"x","b":[1,{"c":null,"d":true}]}`},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := canonical.Transform([]byte(c.in))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != c.want {
				t.Errorf("got  %s\nwant %s", got, c.want)
			}
		})
	}
}

func TestRejectsDuplicateKeys(t *testing.T) {
	if _, err := canonical.Transform([]byte(`{"a":1,"a":2}`)); err == nil {
		t.Error("duplicate keys accepted")
	}
}
