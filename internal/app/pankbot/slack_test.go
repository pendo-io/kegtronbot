package pankbot

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"

	"github.com/pendo-io/appwrap"
	. "gopkg.in/check.v1"
)

type SlackTests struct {
	ctx    context.Context
	log    appwrap.Logging
	closer func()
}

var _ = Suite(&SlackTests{})

func (s *SlackTests) SetUpSuite(c *C) {
	if ctx, closer, err := aetest.NewContext(); err != nil {
		c.FailNow()
	} else {
		s.ctx = ctx
		s.closer = closer
	}

	s.log = appwrap.NewAppengineLogging(s.ctx)
}

func (s *SlackTests) TearDownSuite(c *C) {
	s.closer()
}
