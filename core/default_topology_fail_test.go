package core

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"sync"
	"testing"
)

type stubInitTerminateBoxSharedConfig struct {
	initFailAt int
	initCnt    int

	terminateFailAt  int
	terminatePanicAt int
	terminateCnt     int
}

type stubInitTerminateBox struct {
	*terminateChecker
	init struct {
		called bool
		block  bool
		failed bool
		wg     sync.WaitGroup
	}

	shared *stubInitTerminateBoxSharedConfig
}

var _ StatefulBox = &stubInitTerminateBox{}

func newStubInitTerminateBox(b Box, s *stubInitTerminateBoxSharedConfig) *stubInitTerminateBox {
	return &stubInitTerminateBox{
		terminateChecker: newTerminateChecker(b),
		shared:           s,
	}
}

func (s *stubInitTerminateBox) Init(ctx *Context) error {
	i := &s.init
	i.called = true
	if i.block {
		i.wg.Add(1)
		i.wg.Wait()
	}
	s.shared.initCnt++
	if s.shared.initCnt == s.shared.initFailAt {
		i.failed = true
		return fmt.Errorf("failure")
	}
	return nil
}

func (s *stubInitTerminateBox) ResumeInit() {
	s.init.wg.Done()
}

func (s *stubInitTerminateBox) Terminate(ctx *Context) error {
	s.shared.terminateCnt++
	if err := s.terminateChecker.Terminate(ctx); err != nil {
		return err
	}
	if s.shared.terminateCnt == s.shared.terminatePanicAt {
		panic(fmt.Errorf("failure"))
	}
	if s.shared.terminateCnt == s.shared.terminateFailAt {
		return fmt.Errorf("failure")
	}
	return nil
}

type panicBox struct {
	ProxyBox

	m            sync.Mutex
	writeFailAt  int
	writePanicAt int
	writeCnt     int
}

func (b *panicBox) Process(ctx *Context, t *Tuple, w Writer) error {
	b.m.Lock()
	defer b.m.Unlock()
	b.writeCnt++
	if b.writeCnt == b.writePanicAt {
		panic(fmt.Errorf("test failure via panic"))
	}
	if b.writeCnt == b.writeFailAt {
		return fmt.Errorf("test failure")
	}
	return b.ProxyBox.Process(ctx, t, w)
}

func TestDefaultTopologyFailure(t *testing.T) {
	Convey("Given a simple linear topology", t, func() {
		/*
		 *   so -*--> b1 -*--> si
		 */
		dt := NewDefaultTopology(NewContext(nil), "dt1")
		t := dt.(*defaultTopology)
		Reset(func() {
			t.Stop()
		})

		so := NewTupleIncrementalEmitterSource(freshTuples())
		_, err := t.AddSource("source", so, nil)
		So(err, ShouldBeNil)

		b1 := &panicBox{
			ProxyBox: ProxyBox{
				b: &BlockingForwardBox{cnt: 8},
			},
		}
		tc1 := newTerminateChecker(b1)
		bn1, err := t.AddBox("box1", tc1, nil)
		So(err, ShouldBeNil)
		So(bn1.Input("source", nil), ShouldBeNil)

		si := NewTupleCollectorSink()
		sic := &sinkCloseChecker{s: si}
		sin, err := t.AddSink("sink", sic, nil)
		So(err, ShouldBeNil)
		So(sin.Input("box1", nil), ShouldBeNil)

		Convey("When a box panics", func() {
			b1.writePanicAt = 1
			so.EmitTuples(5)

			Convey("Then the box stops", func() {
				So(bn1.State().Wait(TSStopped), ShouldEqual, TSStopped)
			})

			Convey("Then the topology can be recovered by manual connection", func() {
				So(sin.Input("source", nil), ShouldBeNil)
				so.EmitTuples(3)
				si.Wait(3)
				So(len(si.Tuples), ShouldEqual, 3)
			})
		})

		Convey("When adding a new source with the duplicated name", func() {
			_, err := t.AddSource("SOURCE", so, nil)

			Convey("Then it should fail", func() {
				So(err.Error(), ShouldContainSubstring, "already used")
			})
		})

		Convey("When adding a new sink with the duplicated name", func() {
			_, err := t.AddSink("SINK", sic, nil)

			Convey("Then it should fail", func() {
				So(err.Error(), ShouldContainSubstring, "already used")
			})
		})

		Convey("When adding a new box with the duplicated name", func() {
			_, err := t.AddBox("BOX1", tc1, nil)

			Convey("Then it should fail", func() {
				So(err.Error(), ShouldContainSubstring, "already used")
			})
		})

		Convey("When adding a new input with the duplicated name", func() {
			err := sin.Input("BOX1", nil)

			Convey("Then it should fail", func() {
				So(err.Error(), ShouldContainSubstring, "already")
			})
		})

		// TODO: add more fail tests!!
	})
}
