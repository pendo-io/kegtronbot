package pankbot

import (
	. "gopkg.in/check.v1"
)

type UtilTests struct{}

var _ = Suite(&UtilTests{})

func (s *UtilTests) TestMapListKey(c *C) {
	c.Assert(mapListKey([]string{"a", "b"}), DeepEquals, map[string]bool{"a": true, "b": true})
}

func (s *UtilTests) TestMapKeyList(c *C) {
	c.Assert(mapKeyList(map[string]bool{"a": true, "b": true}), DeepEquals, []string{"a", "b"})
}

func (s *UtilTests) TestUnique(c *C) {
	c.Assert(unique([]string{"c", "b", "a", "a", "d"}), DeepEquals, []string{"a", "b", "c", "d"})
}

func (s *UtilTests) TestStringDifference(c *C) {
	c.Assert(stringDifference([]string{"keep", "remove", "this", "one"}, []string{"remove"}), DeepEquals, []string{"keep", "this", "one"})
	c.Assert(stringDifference([]string{"keep", "remove", "this", "remove", "remove", "one"}, []string{"remove"}), DeepEquals, []string{"keep", "this", "one"})
	c.Assert(stringDifference([]string{"keep", "this"}, []string{}), DeepEquals, []string{"keep", "this"})
	c.Assert(stringDifference([]string{}, []string{"ignore"}), DeepEquals, []string{})
	c.Assert(stringDifference([]string{}, []string{}), DeepEquals, []string{})
}

func (s *UtilTests) TestStringIntersection(c *C) {
	c.Assert(stringIntersection([]string{"a", "b", "c"}, []string{"b"}), DeepEquals, []string{"b"})
	c.Assert(stringIntersection([]string{"b"}, []string{"b", "a", "c"}), DeepEquals, []string{"b"})
	c.Assert(stringIntersection([]string{"a"}, []string{"b"}), DeepEquals, []string{})
	c.Assert(stringIntersection([]string{}, []string{}), DeepEquals, []string{})
}

func (s *UtilTests) TestMinMax(c *C) {
	c.Assert(min(1, 2), Equals, 1)
	c.Assert(min(2, 1), Equals, 1)
	c.Assert(max(1, 2), Equals, 2)
	c.Assert(max(2, 1), Equals, 2)
}

func (s *UtilTests) TestStringSliceContains(c *C) {
	mySlice := []string{"abc", "def", "hij"}

	c.Check(stringSliceContains(mySlice, "abc"), Equals, true)
	c.Check(stringSliceContains(mySlice, "xyz"), Equals, false)
	c.Check(stringSliceContains(mySlice, ""), Equals, false)

	emptySlice := []string{}

	c.Check(stringSliceContains(emptySlice, "abc"), Equals, false)
	c.Check(stringSliceContains(emptySlice, ""), Equals, false)
}

func (s *UtilTests) TestStringSliceToDelimitedString(c *C) {
	singleItemSlice := []string{"abc"}
	c.Check(stringSliceToDelimitedString(singleItemSlice, "|"), Equals, "abc")

	mySlice := []string{"abc", "def", "hij"}

	c.Check(stringSliceToDelimitedString(mySlice, "|"), Equals, "abc|def|hij")
	c.Check(stringSliceToDelimitedString(mySlice, ", "), Equals, "abc, def, hij")

	emptySlice := []string{}

	c.Check(stringSliceToDelimitedString(emptySlice, "|"), Equals, "")
}

type exactJsonTester struct {
	StringField   string   `json:"stringV"`
	IntField      int      `json:"intV"`
	SliceField    []string `json:"slice"`
	OptionalField string   `json:"optional,omitempty"`
}

func (s *UtilTests) TestUnmarshalExactJson(c *C) {
	// bad JSON
	c.Check(unmarshalExactJson([]byte{}, &exactJsonTester{}), ErrorMatches, ".*JSON input.*")

	// Wrong data type
	c.Check(unmarshalExactJson([]byte(`{"stringV":"foo","intV":23.5,"slice":[]}`), &exactJsonTester{}), ErrorMatches, ".*cannot unmarshal number.*")

	// Valid, omitempty field dropped
	c.Check(unmarshalExactJson([]byte(`{"stringV":"foo","intV":23,"slice":[]}`), &exactJsonTester{}), IsNil)
	// Whitespace doesn't matter
	c.Check(unmarshalExactJson([]byte(`{"stringV":  "foo",
	"intV":23,	"slice":[    ]}`), &exactJsonTester{}), IsNil)
	// Omitempty field included
	c.Check(unmarshalExactJson([]byte(`{"stringV":"foo","intV":23,"slice":[],"optional":"bar"}`), &exactJsonTester{}), IsNil)
	// All zero values
	c.Check(unmarshalExactJson([]byte(`{"stringV":"","intV":0,"slice":[]}`), &exactJsonTester{}), IsNil)
	// TODO: support zero values in omitempty fields
	//c.Check(unmarshalExactJson([]byte(`{"stringV":"foo","intV":23,"slice":[],"optional":""}`), &exactJsonTester{}), IsNil)

	// Missing required field(s)
	c.Check(unmarshalExactJson([]byte(`{}`), &exactJsonTester{}), ErrorMatches, ".*fields do not match target.*")
	c.Check(unmarshalExactJson([]byte(`{"stringV":"foo","intV":23}`), &exactJsonTester{}), ErrorMatches, ".*fields do not match target.*")
	c.Check(unmarshalExactJson([]byte(`{"stringV":"foo","slice":[]}`), &exactJsonTester{}), ErrorMatches, ".*fields do not match target.*")
	c.Check(unmarshalExactJson([]byte(`{"intV":23,"slice":[]}`), &exactJsonTester{}), ErrorMatches, ".*fields do not match target.*")

	// Extra field(s)
	c.Check(unmarshalExactJson([]byte(`{"stringV":"foo","intV":23,"slice":[],"extra":"baz"}`), &exactJsonTester{}), ErrorMatches, ".*fields do not match.*")
}
