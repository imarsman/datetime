// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"

	"github.com/imarsman/datetime/period"
	. "github.com/onsi/gomega"
)

func TestGobEncoding(t *testing.T) {
	g := NewGomegaWithT(t)

	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	decoder := gob.NewDecoder(&b)
	cases := []string{
		"P0D",
		"P1D",
		"P7D",
		"P1M",
		"P1Y",
		"PT1H",
		"PT1M",
		"PT1S",
		"P2Y3M4W5D",
		"-P2Y3M4W5D",
		"P2Y3M4W5DT1H7M9S",
		"-P2Y3M4W5DT1H7M9S",
		"P48M",
	}
	for i, c := range cases {
		p := period.MustParse(c, false)
		// var p2 period.Period
		var p2 period.Period
		err := encoder.Encode(&p)
		// fmt.Println(p)
		g.Expect(err).NotTo(HaveOccurred(), info(i, c))
		if err == nil {
			err = decoder.Decode(&p2)
			g.Expect(err).NotTo(HaveOccurred(), info(i, c))
			g.Expect(p).To(Equal(p2), info(i, c))
		}
	}
}

func TestPeriodJSONMarshalling(t *testing.T) {
	g := NewGomegaWithT(t)

	cases := []struct {
		value period.Period
		want  string
	}{
		{period.New(-1111, -4, -3, -11, -59, -59), `"-P1111Y4M3DT11H59M59S"`},
		{period.New(-1, -10, -31, -5, -4, -20), `"-P1Y10M31DT5H4M20S"`},
		{period.New(0, 0, 0, 0, 0, 0), `"P0D"`},
		{period.New(0, 0, 0, 0, 0, 1), `"PT1S"`},
		{period.New(0, 0, 0, 0, 1, 0), `"PT1M"`},
		{period.New(0, 0, 0, 1, 0, 0), `"PT1H"`},
		{period.New(0, 0, 1, 0, 0, 0), `"P1D"`},
		{period.New(0, 1, 0, 0, 0, 0), `"P1M"`},
		{period.New(1, 0, 0, 0, 0, 0), `"P1Y"`},
	}
	for i, c := range cases {
		var p period.Period
		bb, err := json.Marshal(c.value)
		g.Expect(err).NotTo(HaveOccurred(), info(i, c))
		g.Expect(string(bb)).To(Equal(c.want), info(i, c))
		if string(bb) == c.want {
			err = json.Unmarshal(bb, &p)
			g.Expect(err).NotTo(HaveOccurred(), info(i, c))
			g.Expect(p).To(Equal(c.value), info(i, c))
		}
	}
}

func TestPeriodTextMarshalling(t *testing.T) {
	g := NewGomegaWithT(t)

	cases := []struct {
		value period.Period
		want  string
	}{
		{period.New(-1111, -4, -3, -11, -59, -59), "-P1111Y4M3DT11H59M59S"},
		{period.New(-1, -9, -31, -5, -4, -20), "-P1Y9M31DT5H4M20S"},
		{period.New(0, 0, 0, 0, 0, 0), "P0D"},
		{period.New(0, 0, 0, 0, 0, 1), "PT1S"},
		{period.New(0, 0, 0, 0, 1, 0), "PT1M"},
		{period.New(0, 0, 0, 1, 0, 0), "PT1H"},
		{period.New(0, 0, 1, 0, 0, 0), "P1D"},
		{period.New(0, 1, 0, 0, 0, 0), "P1M"},
		{period.New(1, 0, 0, 0, 0, 0), "P1Y"},
	}
	for i, c := range cases {
		var p period.Period
		c.value.Input = c.want
		bb, err := c.value.MarshalText()
		// fmt.Println("bytes", string(bb))
		g.Expect(err).NotTo(HaveOccurred(), info(i, c))
		g.Expect(string(bb)).To(Equal(c.want), info(i, c))
		if string(bb) == c.want {
			err = p.UnmarshalText(bb)
			// fmt.Println("c", c)
			g.Expect(err).NotTo(HaveOccurred(), info(i, c))
			g.Expect(p).To(Equal(c.value), info(i, c))
		}
	}
}

// func TestInvalidPeriodText(t *testing.T) {
// 	g := NewGomegaWithT(t)

// 	cases := []struct {
// 		value string
// 		want  string
// 	}{
// 		{``, `cannot parse a blank string as a period`},
// 		{`not-a-period`, `not-a-period: expected 'P' period mark at the start`},
// 		{`P000`, `P000: missing designator at the end`},
// 	}
// 	for i, c := range cases {
// 		var p period.Period
// 		err := p.UnmarshalText([]byte(c.value))
// 		g.Expect(err).To(HaveOccurred(), info(i, c))
// 		g.Expect(err.Error()).To(Equal(c.want), info(i, c))
// 	}
// }
