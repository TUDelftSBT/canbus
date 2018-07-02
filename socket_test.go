// Copyright 2016 The go-daq Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package canbus_test

import (
	"crypto/rand"
	"fmt"
	"github.com/TUDelftSBT/canbus"
	"reflect"
	"testing"
)

func setupSocket(dev string) (*canbus.Socket, error) {
	r, err := canbus.New()
	if err != nil {
		return nil, err
	}
	err = r.Bind(dev)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func teardownSocket(s *canbus.Socket) error {
	if err := s.Close(); err != nil {
		return err
	}
	return nil
}

func TestExtendenFrameFormat(t *testing.T) {
	const (
		endpoint = "vcan0"
		ID       = 4096
	)
	msg := []byte{0x13, 0x37}

	r, err := setupSocket(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	w, err := setupSocket(endpoint)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		_, err := w.Send(ID, msg)
		if err != nil {
			t.Fatalf("error send: %v\n", err)
		}
	}()

	id, data, err := r.Recv()
	if id != ID {
		t.Errorf("got id=%v. want=%d\n", id, ID)
	}
	if !reflect.DeepEqual(data[:], msg[:]) {
		t.Errorf("error data:\ngot= %v\nwant=%v\n", data[1:], msg[1:])
	}

	if err := teardownSocket(r); err != nil {
		t.Fatal(err)
	}
	if err := teardownSocket(w); err != nil {
		t.Fatal(err)
	}
}

func TestSocket(t *testing.T) {
	const (
		endpoint = "vcan0"
		N        = 10
		ID       = 128
	)
	var msg = []byte{0, 0xde, 0xad, 0xbe, 0xef}

	r, err := setupSocket(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	w, err := setupSocket(endpoint)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Recv", func(t *testing.T) {
		go func() {
			for i := 0; i < N; i++ {
				msg[0] = byte(i)
				_, err := w.Send(ID, msg)
				if err != nil {
					t.Fatalf("error send[%d]: %v\n", i, err)
				}

			}
		}()

		for i := 0; i < N; i++ {
			id, data, err := r.Recv()
			if err != nil {
				t.Fatalf("error recv: %v\n", err)
			}
			if id != ID {
				t.Errorf("got id=%v. want=%d\n", id, ID)
			}
			if got, want := int(data[0]), i; got != want {
				t.Errorf("error index: got=%d. want=%d\n", got, want)
			}

			if !reflect.DeepEqual(data[1:], msg[1:]) {
				t.Errorf("error data:\ngot= %v\nwant=%v\n", data[1:], msg[1:])
			}
		}
	})

	t.Run("RecvRaw", func(t *testing.T) {
		go func() {
			for i := 0; i < N; i++ {
				msg[0] = byte(i)
				_, err := w.Send(ID, msg)
				if err != nil {
					t.Fatalf("error send[%d]: %v\n", i, err)
				}

			}
		}()
		canmsg := canbus.CANMessage{}

		for i := 0; i < N; i++ {
			err := r.RecvRaw(&canmsg)

			id := canmsg.GetID()
			can_dlc := canmsg.GetLen()
			if err != nil {
				t.Fatalf("error recv: %v\n", err)
			}
			if id != ID {
				t.Errorf("got id=%v. want=%d\n", id, ID)
			}

			data := canmsg.GetData()
			if got, want := int(data[0]), i; got != want {
				t.Errorf("error index: got=%d. want=%d\n", got, want)
			}

			if !reflect.DeepEqual(data[1:can_dlc], msg[1:]) {
				t.Errorf("error data:\ngot= %v\nwant=%v\n", data[1:], msg[1:])
			}

		}

	})
	if err := teardownSocket(r); err != nil {
		t.Fatal(err)
	}
	if err := teardownSocket(w); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkRecv(b *testing.B) {
	const (
		dev = "vcan0"
		ID  = 707
	)

	//Setup
	r, err := setupSocket(dev)
	if err != nil {
		b.Fatal(err)
	}
	w, err := setupSocket(dev)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()

	canmsg := canbus.CANMessage{}
	quit := make(chan bool)
	sendmsgs := func(l int) {
		msg := make([]byte, l)
		_, err := rand.Read(msg)
		if err != nil {
			b.Fatalf("Error reading random bytes: %v\n", err)
		}
		for {
			select {
			case <-quit:
				return
			default:
				_, err := w.Send(ID, msg)
				if err != nil {
					b.Fatalf("error send: %v\n", err)
				}
			}

		}
	}

	for l := 2; l <= 8; l = l * 2 {
		go sendmsgs(l)
		b.Run(fmt.Sprintf("RecvRaw%db", l), func(b *testing.B) {
			b.SetBytes(int64(l))
			for i := 0; i < b.N; i++ {
				err := r.RecvRaw(&canmsg)
				if err != nil {
					b.Fatalf("error RecvRaw: %v\n", err)
				}
			}

		})
		b.Run(fmt.Sprintf("Recv%db", l), func(b *testing.B) {
			b.SetBytes(int64(l))
			for i := 0; i < b.N; i++ {
				_, _, err := r.Recv()
				if err != nil {
					b.Fatalf("error Recv: %v\n", err)
				}
			}

		})
		quit <- true

	}
	teardownSocket(r)
	teardownSocket(w)
}
