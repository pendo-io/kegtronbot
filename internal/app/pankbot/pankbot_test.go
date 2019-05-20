package pankbot

import (
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"

	"github.com/pendo-io/appwrap"
	. "gopkg.in/check.v1"
)

type PankTests struct {
	ctx    context.Context
	log    appwrap.Logging
	closer func()
}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&PankTests{})

func (s *PankTests) SetUpSuite(c *C) {
	if ctx, closer, err := aetest.NewContext(); err != nil {
		c.FailNow()
	} else {
		s.ctx = ctx
		s.closer = closer
	}
	s.log = appwrap.NullLogger{}
}

func (s *PankTests) TearDownSuite(c *C) {
	s.closer()
}

func (s *PankTests) TestPrivatePank(c *C) {
	c.Check(isPrivatePank([]string{"foo", "because"}), Equals, false)
	c.Check(isPrivatePank([]string{"foo", "private"}), Equals, true)
	c.Check(isPrivatePank([]string{"foo", "privately"}), Equals, true)
	c.Check(isPrivatePank([]string{"foo", "private(ly)"}), Equals, true)
	c.Check(isPrivatePank([]string{"foo", "<private(ly)<"}), Equals, true)
	c.Check(isPrivatePank([]string{"foo", "<private(ly)>"}), Equals, true)
	c.Check(isPrivatePank([]string{"foo", "<private<"}), Equals, true)
}

func (s *PankTests) TestQuarterStart(c *C) {
	loc, _ := time.LoadLocation("America/New_York")
	// check inside each month
	Jan2 := time.Date(2016, 1, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Jan2.UTC()), DeepEquals, time.Date(2015, 11, 1, 0, 0, 0, 0, loc).UTC())
	Feb2 := time.Date(2016, 2, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Feb2.UTC()), DeepEquals, time.Date(2016, 2, 1, 0, 0, 0, 0, loc).UTC())
	Mar2 := time.Date(2016, 3, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Mar2.UTC()), DeepEquals, time.Date(2016, 2, 1, 0, 0, 0, 0, loc).UTC())
	Apr2 := time.Date(2016, 4, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Apr2.UTC()), DeepEquals, time.Date(2016, 2, 1, 0, 0, 0, 0, loc).UTC())
	May2 := time.Date(2016, 5, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(May2.UTC()), DeepEquals, time.Date(2016, 5, 1, 0, 0, 0, 0, loc).UTC())
	Jun2 := time.Date(2016, 6, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Jun2.UTC()), DeepEquals, time.Date(2016, 5, 1, 0, 0, 0, 0, loc).UTC())
	Jul2 := time.Date(2016, 7, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Jul2.UTC()), DeepEquals, time.Date(2016, 5, 1, 0, 0, 0, 0, loc).UTC())
	Aug2 := time.Date(2016, 8, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Aug2.UTC()), DeepEquals, time.Date(2016, 8, 1, 0, 0, 0, 0, loc).UTC())
	Sep2 := time.Date(2016, 9, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Sep2.UTC()), DeepEquals, time.Date(2016, 8, 1, 0, 0, 0, 0, loc).UTC())
	Oct2 := time.Date(2016, 10, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Oct2.UTC()), DeepEquals, time.Date(2016, 8, 1, 0, 0, 0, 0, loc).UTC())
	Nov2 := time.Date(2016, 11, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Nov2.UTC()), DeepEquals, time.Date(2016, 11, 1, 0, 0, 0, 0, loc).UTC())
	Dec2 := time.Date(2016, 12, 2, 3, 4, 5, 6, loc)
	c.Check(quarterStart(Dec2.UTC()), DeepEquals, time.Date(2016, 11, 1, 0, 0, 0, 0, loc).UTC())

	// check some boundaries
	Dec31 := time.Date(2015, 12, 31, 23, 59, 59, 99999999, loc)
	c.Check(quarterStart(Dec31), DeepEquals, time.Date(2015, 11, 1, 0, 0, 0, 0, loc).UTC())
	c.Check(quarterStart(Dec31.UTC()), DeepEquals, time.Date(2015, 11, 1, 0, 0, 0, 0, loc).UTC())
	Jan1 := time.Date(2016, 1, 1, 1, 0, 0, 0, loc)
	c.Check(quarterStart(Jan1), DeepEquals, time.Date(2015, 11, 1, 0, 0, 0, 0, loc).UTC())
	c.Check(quarterStart(Jan1.UTC()), DeepEquals, time.Date(2015, 11, 1, 0, 0, 0, 0, loc).UTC())
	Jan31 := time.Date(2016, 1, 31, 23, 59, 59, 99999999, loc)
	c.Check(quarterStart(Jan31), DeepEquals, time.Date(2015, 11, 1, 0, 0, 0, 0, loc).UTC())
	c.Check(quarterStart(Jan31.UTC()), DeepEquals, time.Date(2015, 11, 1, 0, 0, 0, 0, loc).UTC())
	Feb1 := time.Date(2016, 2, 1, 0, 0, 0, 0, loc)
	c.Check(quarterStart(Feb1), DeepEquals, time.Date(2016, 2, 1, 0, 0, 0, 0, loc).UTC())
	c.Check(quarterStart(Feb1.UTC()), DeepEquals, time.Date(2016, 2, 1, 0, 0, 0, 0, loc).UTC())
}

func (s *PankTests) TestMonthStart(c *C) {
	loc, _ := time.LoadLocation("America/New_York")
	// check inside each month
	Jan2 := time.Date(2016, 1, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Jan2.UTC()), DeepEquals, time.Date(2016, 1, 1, 0, 0, 0, 0, loc).UTC())
	Feb2 := time.Date(2016, 2, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Feb2.UTC()), DeepEquals, time.Date(2016, 2, 1, 0, 0, 0, 0, loc).UTC())
	Mar2 := time.Date(2016, 3, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Mar2.UTC()), DeepEquals, time.Date(2016, 3, 1, 0, 0, 0, 0, loc).UTC())
	Apr2 := time.Date(2016, 4, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Apr2.UTC()), DeepEquals, time.Date(2016, 4, 1, 0, 0, 0, 0, loc).UTC())
	May2 := time.Date(2016, 5, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(May2.UTC()), DeepEquals, time.Date(2016, 5, 1, 0, 0, 0, 0, loc).UTC())
	Jun2 := time.Date(2016, 6, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Jun2.UTC()), DeepEquals, time.Date(2016, 6, 1, 0, 0, 0, 0, loc).UTC())
	Jul2 := time.Date(2016, 7, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Jul2.UTC()), DeepEquals, time.Date(2016, 7, 1, 0, 0, 0, 0, loc).UTC())
	Aug2 := time.Date(2016, 8, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Aug2.UTC()), DeepEquals, time.Date(2016, 8, 1, 0, 0, 0, 0, loc).UTC())
	Sep2 := time.Date(2016, 9, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Sep2.UTC()), DeepEquals, time.Date(2016, 9, 1, 0, 0, 0, 0, loc).UTC())
	Oct2 := time.Date(2016, 10, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Oct2.UTC()), DeepEquals, time.Date(2016, 10, 1, 0, 0, 0, 0, loc).UTC())
	Nov2 := time.Date(2016, 11, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Nov2.UTC()), DeepEquals, time.Date(2016, 11, 1, 0, 0, 0, 0, loc).UTC())
	Dec2 := time.Date(2016, 12, 2, 3, 4, 5, 6, loc)
	c.Check(monthStart(Dec2.UTC()), DeepEquals, time.Date(2016, 12, 1, 0, 0, 0, 0, loc).UTC())

	// check some boundaries
	Dec31 := time.Date(2015, 12, 31, 23, 59, 59, 99999999, loc)
	c.Check(monthStart(Dec31), DeepEquals, time.Date(2015, 12, 1, 0, 0, 0, 0, loc).UTC())
	c.Check(monthStart(Dec31.UTC()), DeepEquals, time.Date(2015, 12, 1, 0, 0, 0, 0, loc).UTC())
	Jan1 := time.Date(2016, 1, 1, 1, 0, 0, 0, loc)
	c.Check(monthStart(Jan1), DeepEquals, time.Date(2016, 1, 1, 0, 0, 0, 0, loc).UTC())
	c.Check(monthStart(Jan1.UTC()), DeepEquals, time.Date(2016, 1, 1, 0, 0, 0, 0, loc).UTC())
	Jan31 := time.Date(2016, 1, 31, 23, 59, 59, 99999999, loc)
	c.Check(monthStart(Jan31), DeepEquals, time.Date(2016, 1, 1, 0, 0, 0, 0, loc).UTC())
	c.Check(monthStart(Jan31.UTC()), DeepEquals, time.Date(2016, 1, 1, 0, 0, 0, 0, loc).UTC())
	Feb1 := time.Date(2016, 2, 1, 0, 0, 0, 0, loc)
	c.Check(monthStart(Feb1), DeepEquals, time.Date(2016, 2, 1, 0, 0, 0, 0, loc).UTC())
	c.Check(monthStart(Feb1.UTC()), DeepEquals, time.Date(2016, 2, 1, 0, 0, 0, 0, loc).UTC())
}
