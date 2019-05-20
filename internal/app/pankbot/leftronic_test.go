package pankbot

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"

	"github.com/pendo-io/appwrap"
	. "gopkg.in/check.v1"
)

type LeftronicTests struct {
	ctx    context.Context
	log    appwrap.Logging
	closer func()
}

var _ = Suite(&LeftronicTests{})

func (s *LeftronicTests) SetUpSuite(c *C) {
	if ctx, closer, err := aetest.NewContext(); err != nil {
		c.FailNow()
	} else {
		s.ctx = ctx
		s.closer = closer
	}
	s.log = appwrap.NewAppengineLogging(s.ctx)
	leftronicUrl = "https://localhost/customSend"
}

func (s *LeftronicTests) TearDownSuite(c *C) {
	s.closer()
}

func (s *LeftronicTests) TestPayload(c *C) {
	l := Leftronic("access key", "stream name", "<html>")
	c.Check(l.AccessKey, Equals, "access key")
	c.Check(l.StreamName, Equals, "stream name")
	c.Check(l.Point.Html, Equals, "<html>")
}

func (s *LeftronicTests) TestPost(c *C) {
	l := Leftronic("access key", "stream name", "<html>")
	err := l.Post(s.ctx)
	c.Assert(err, NotNil)
}
