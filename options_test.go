package xbot

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type BotOptsSuite struct {
	suite.Suite
	opt BotOpts
}

func TestBotCfg(t *testing.T) {
	suite.Run(t, new(BotOptsSuite))
}

func (s *BotOptsSuite) SetupSuite() {
	s.opt.BotCfg = &BotConfig{}
}

func (s *BotOptsSuite) Test_00_Default() {
	bo := s.opt
	s.Equal(bo.BotCfg.Screen, 0)
	s.Equal(bo.BotCfg.Headless, false)
	s.Equal(bo.ElemIndex, 0)
	s.Equal(bo.Attr, "")
	s.Equal(bo.Property, "")
	s.Equal(bo.Submit, false)
}

func (s *BotOptsSuite) setParams(opts ...BotOptFunc) {
	for _, f := range opts {
		f(&s.opt)
	}
}

func (s *BotOptsSuite) Test_01_BotScreen() {
	f := BotScreen(1)

	tp := reflect.TypeOf(f)
	s.Equal(tp.String(), "xbot.BotOptFunc")
	s.Equal(tp.Kind(), reflect.Func)

	s.setParams(f)
	s.Equal(s.opt.BotCfg.Screen, 1)
}

func (s *BotOptsSuite) Test_02_BotHeadless() {
	f := BotHeadless(true)
	s.setParams(f)
	s.Equal(s.opt.BotCfg.Headless, true)
}

func (s *BotOptsSuite) Test_03_BotElemIndex() {
	f := ElemIndex(1)
	s.setParams(f)
	s.Equal(s.opt.ElemIndex, 1)
}

func (s *BotOptsSuite) Test_04_BotElemAttr() {
	f := ElemAttr("href")
	s.setParams(f)
	s.Equal(s.opt.Attr, "href")
}

func (s *BotOptsSuite) Test_05_BotElemProp() {
	v := "value"
	f := ElemProperty(v)
	s.setParams(f)
	s.Equal(s.opt.Property, v)
}

func (s *BotOptsSuite) Test_06_BotInputSubmit() {
	v := true
	f := InputSubmit(v)
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
		f := BotHighlight(tt.args.b)
		s.setParams(f)
		s.Equal(tt.want, s.opt.BotCfg.Highlight, tt.name)
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
		f := BotSteps(tt.args.i)
		s.setParams(f)
		s.Equal(tt.want, s.opt.BotCfg.Steps, tt.name)
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
		fn := ElemOffsetToTop(tt.args.f)
		s.setParams(fn)
		s.Equal(tt.want, s.opt.OffsetToTop, tt.name)
	}
}

func (s *BotOptsSuite) TestBotTimeout() {
	type args struct {
		i int
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
		f := BotTimeout(tt.args.i)
		s.setParams(f)
		s.Equal(tt.want, s.opt.Timeout, tt.want)
	}
}
