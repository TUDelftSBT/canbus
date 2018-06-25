// Copyright 2016 The go-daq Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package canbus_test

import (
	"github.com/TUDelftSBT/canbus"
	"log"
	"reflect"
	"testing"
)

func setupSocket(t *testing.T, dev string) *canbus.Socket {
	r, err := canbus.New()
	if err != nil {
		t.Fatal(err)
	}
	err = r.Bind(dev)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func teardownSocket(t *testing.T, s *canbus.Socket) {
	if err := s.Close(); err != nil {
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

	r := setupSocket(t, endpoint)
	w := setupSocket(t, endpoint)

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

	teardownSocket(t, r)
	teardownSocket(t, w)

}
