package xbot_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/coghost/xbot"
	"github.com/stretchr/testify/suite"
)

type BotOptsSuite struct {
	suite.Suite
	opt xbot.BotOpts
}

func TestBotCfg(t *testing.T) {
	suite.Run(t, new(BotOptsSuite))
}

func (s *BotOptsSuite) SetupSuite() {
}

func (s *BotOptsSuite) Test_00_Default() {
	bo := s.opt
	s.Equal(bo.Screen, 0)
	s.Equal(bo.Headless, false)
	s.Equal(bo.ElemIndex, 0)
	s.Equal(bo.Attr, "")
	s.Equal(bo.Property, "")
	s.Equal(bo.Submit, false)
}

func (s *BotOptsSuite) setParams(opts ...xbot.BotOptFunc) {
	for _, f := range opts {
		f(&s.opt)
	}
}

func (s *BotOptsSuite) Test_01_BotScreen() {
	f := xbot.BotScreen(1)

	tp := reflect.TypeOf(f)
	s.Equal(tp.String(), "xbot.BotOptFunc")
	s.Equal(tp.Kind(), reflect.Func)

	s.setParams(f)
	s.Equal(s.opt.Screen, 1)
}

func (s *BotOptsSuite) Test_02_BotHeadless() {
	f := xbot.BotHeadless(true)
	s.setParams(f)
	s.Equal(s.opt.Headless, true)
}

func (s *BotOptsSuite) Test_03_BotElemIndex() {
	f := xbot.ElemIndex(1)
	s.setParams(f)
	s.Equal(s.opt.ElemIndex, 1)
}

func (s *BotOptsSuite) Test_04_BotElemAttr() {
	f := xbot.ElemAttr("href")
	s.setParams(f)
	s.Equal(s.opt.Attr, "href")
}

func (s *BotOptsSuite) Test_05_BotElemProp() {
	v := "value"
	f := xbot.ElemProperty(v)
	s.setParams(f)
	s.Equal(s.opt.Property, v)
}

func (s *BotOptsSuite) Test_06_BotInputSubmit() {
	v := true
	f := xbot.InputSubmit(v)
	s.setParams(f)
	s.Equal(s.opt.Submit, v)
}

func (s *BotOptsSuite) TestBotHighlight() {
	type args struct {
		b bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test with true", args: args{b: true}, want: true},
		{name: "test with false", args: args{b: false}, want: false},
	}
	for _, tt := range tests {
		f := xbot.BotHighlight(tt.args.b)
		s.setParams(f)
		s.Equal(tt.want, s.opt.Highlight, tt.name)
	}
}

func (s *BotOptsSuite) TestBotSteps() {
	type args struct {
		i int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "default", args: args{i: 0}, want: 0},
	}
	for _, tt := range tests {
		f := xbot.BotSteps(tt.args.i)
		s.setParams(f)
		s.Equal(tt.want, s.opt.Steps, tt.name)
	}
}

func (s *BotOptsSuite) TestElemOffsetToTop() {
	type args struct {
		f float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "default", args: args{f: 0}, want: 0},
		{name: "0.1", args: args{f: 0.1}, want: 0.1},
		{name: "1024.0", args: args{f: 1024.0}, want: 1024.0},
	}
	for _, tt := range tests {
		fn := xbot.ElemOffsetToTop(tt.args.f)
		s.setParams(fn)
		s.Equal(tt.want, s.opt.OffsetToTop, tt.name)
	}
}

func (s *BotOptsSuite) TestBotTimeout() {
	type args struct {
		i time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{name: "default", args: args{i: 0}, want: 0},
		{name: "1", args: args{i: 1}, want: 1},
	}
	for _, tt := range tests {
		f := xbot.BotTimeout(tt.args.i)
		s.setParams(f)
		s.Equal(tt.want, s.opt.Timeout, tt.want)
	}
}
