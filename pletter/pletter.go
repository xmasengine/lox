package pletter

import "io"

/*
  Pletter v0.5c1

  XL2S Entertainment
  Converted from C++ to Go by Xmas Engine
*/

var maxlen = [7]int{128, 128 + 128, 512 + 128, 1024 + 128, 2048 + 128, 4096 + 128, 8192 + 128}
var varcost [65536]int

type metadata struct {
	reeks int
	cpos  [7]int
	clen  [7]int
}

type pakdata struct {
	cost int
	mode int
	mlen int
}

type saves struct {
	SaveLength bool
	length     int
	offset     int
	pak        [7][]pakdata
	m          []metadata
	buf        []byte
	d          []byte
	ep         int
	dp         int
	p          int
	e          int
}

func (s *saves) init() {
	s.ep = 0
	s.dp = 0
	s.p = 0
	s.e = 0
	s.buf = make([]byte, s.length*2)
}

func (s *saves) add0() {
	if s.p == 0 {
		s.claimevent()
	}
	s.e *= 2
	s.p++
	if s.p == 8 {
		s.addevent()
	}
}

func (s *saves) add1() {
	if s.p == 0 {
		s.claimevent()
	}
	s.e *= 2
	s.p++
	s.e++
	if s.p == 8 {
		s.addevent()
	}
}
func (s *saves) addbit(b int) {
	if b > 0 {
		s.add1()
	} else {
		s.add0()
	}
}
func (s *saves) add3(b int) {
	s.addbit(b & 4)
	s.addbit(b & 2)
	s.addbit(b & 1)
}
func (s *saves) addVar(i int) {
	for j := 32768; (i & j) > 0; j /= 2 {
		if j == 1 {
			s.add0()
		}
		j /= 2
		s.add1()
		s.addbit(i & j)
	}
}
func (s *saves) addData(d byte) {
	s.dp++
	s.buf[s.dp] = d
}

func (s *saves) addevent() {
	s.buf[s.ep] = byte(s.e)
	s.e = 0
	s.p = 0
}

func (s *saves) claimevent() {
	s.ep = s.dp
	s.dp++
}

func (s *saves) done() {
	if s.p != 0 {
		for s.p != 8 {
			s.e *= 2
			s.p++
			s.addevent()
		}
	}
}

func New(saveLength bool) *saves {
	return &saves{SaveLength: saveLength}
}

func (s *saves) Save(wr io.Writer) (int, error) {
	println("save")
	s.createmetadata()
	println("createmateadata ok")

	minlen := s.length * 1000
	minbl := 0

	for i := 1; i < 7; i++ {
		s.pak[i] = make([]pakdata, s.length)
		l := s.getlen(s.pak[i], i)
		if l < minlen && i > 0 {
			minlen = l
			minbl = i
		}
	}
	println("paks ok")
	s.save(s.pak[minbl], minbl)
	println("save ok")
	return wr.Write(s.buf)
}

func (s *saves) Load(rd io.Reader) error {
	buf, err := io.ReadAll(rd)
	if err != nil {
		return err
	}
	s.d = buf
	s.length = len(buf)
	s.m = make([]metadata, s.length)
	s.init()
	return nil
}

// initializes the varcost array
func init() {
	v := 1
	b := 1

	for r := 1; r < len(varcost); r *= 2 {
		for j := 0; j < r; j++ {
			varcost[v] = b
			v++
			if v > len(varcost) {
				break
			}
		}
		b += 2
	}
}

func (s *saves) createmetadata() {
	var i, j int
	last := make([]int, 65536)
	prev := make([]int, s.length+1)

	for i := 0; i < len(last); i++ {
		last[i] = -1
	}

	for i = 0; i < s.length; i++ {
		s.m[i].cpos[0] = 0
		s.m[i].clen[0] = 0
		idx := int(s.d[i])
		if i+1 < s.length {
			idx += int(s.d[i+1]) * 256
		}
		prev[i] = last[idx]
		last[idx] = i
	}
	r := -1
	t := 0
	for i := s.length - 1; i != -1; i-- {
		if s.d[i] == byte(r) {
			t++
			s.m[i].reeks = t
		} else {
			r = int(s.d[i])
			t = 1
			s.m[i].reeks = t
		}
	}
	for bl := 0; bl != 7; bl++ {
		for i = 0; i < s.length; i++ {
			var l int
			var p int
			p = i
			if bl > 0 {
				s.m[i].clen[bl] = s.m[i].clen[bl-1]
				s.m[i].cpos[bl] = s.m[i].cpos[bl-1]
				p = i - s.m[i].cpos[bl]
			}
			for p = prev[p]; p != -1; p = prev[p] {
				if (i - p) > maxlen[bl] {
					break
				}
				l = 0
				for s.d[p+l] == s.d[i+l] && (i+l) < s.length {
					if s.m[i+l].reeks > 1 {
						j = s.m[i+l].reeks
						if j > s.m[p+l].reeks {
							j = s.m[p+l].reeks
							l += j
						} else {
							l++
						}
					}
					if l > s.m[i].clen[bl] {
						s.m[i].clen[bl] = l
						s.m[i].cpos[bl] = i - p
					}
				}
			}
		}
	}
}

func (s *saves) getlen(p []pakdata, q int) int {
	var i, j, cc, ccc, kc, kmode, kl int
	p[s.length].cost = 0
	for i = s.length - 1; i != -1; i-- {
		kmode = 0
		kl = 0
		kc = 9 + p[i+1].cost

		j = s.m[i].clen[0]
		for j > 1 {
			cc = 9 + varcost[j-1] + p[i+j].cost
			if cc < kc {
				kc = cc
				kmode = 1
				kl = j
			}
			j--
		}

		j = s.m[i].clen[q]
		if q == 1 {
			ccc = 9
		} else {
			ccc = 9 + q
		}

		for j > 1 {
			cc = ccc + varcost[j-1] + p[i+j].cost
			if cc < kc {
				kc = cc
				kmode = 2
				kl = j
			}
			j--
		}

		p[i].cost = kc
		p[i].mode = kmode
		p[i].mlen = kl
	}
	return p[0].cost
}

func (s *saves) save(p []pakdata, q int) {
	s.init()
	var i, j int

	if s.SaveLength {
		s.addData(byte(s.length & 255))
		s.addData(byte(s.length >> 8))
	}

	s.add3(q - 1)
	s.addData(s.d[0])

	i = 1
	for i < s.length {
		switch p[i].mode {
		case 0:
			s.add0()
			s.addData(s.d[i])
			i++
		case 1:
			s.add1()
			s.addVar(p[i].mlen - 1)
			j = s.m[i].cpos[0] - 1
			if j > 127 {
				print("-j>128-")
			}
			s.addData(byte(j))
			i += p[i].mlen
		case 2:
			s.add1()
			s.addVar(p[i].mlen - 1)
			j = s.m[i].cpos[q] - 1
			if j < 128 {
				print("-j<128-")
			}
			j -= 128
			s.addData(byte(128 | j&127))
			switch q {
			case 6:
				s.addbit(j & 4096)
				fallthrough
			case 5:
				s.addbit(j & 2048)
				fallthrough
			case 4:
				s.addbit(j & 1024)
				fallthrough
			case 3:
				s.addbit(j & 512)
				fallthrough
			case 2:
				s.addbit(j & 256)
				s.addbit(j & 128)
				fallthrough
			case 1:
				break
			default:
				print("-2-")
				break
			}
			i += p[i].mlen
			break
		default:
			print("-?-")
			break
		}
	}

	for i = 0; i != 34; i++ {
		s.add1()
	}
	s.done()
}


func getbit(mem []int, hl *int) int {
	  a := mem[*hl]
	  (*hl)++
	  a = a<<1
	  return a
}

func GETBIT(mem []int, a*, hl *int) int {
	(*a) = (*a) * 2
	getbit(hl)

func Unpack(in []buf) {

	/*

	; pletter v0.5c msx unpacker

	; call unpack with hl pointing to some pletter5 data, and de pointing to the destination.
	; changes all registers

	; define lengthindata when the original size is written in the pletter data

	;  define LENGTHINDATA

	  module pletter

	  macro GETBIT
	  add a,a
	  call z,getbit
	  endmacro

	  macro GETBITEXX
	  add a,a
	  call z,getbitexx
	  endmacro

	@unpack

	  ifdef LENGTHINDATA
	  inc hl
	  inc hl
	  endif

	  ld a,(hl)
	  inc hl
	  exx
	  ld de,0
	  add a,a
	  inc a
	  rl e
	  add a,a
	  rl e
	  add a,a
	  rl e
	  rl e
	  ld hl,modes
	  add hl,de
	  ld e,(hl)
	  ld ixl,e
	  inc hl
	  ld e,(hl)
	  ld ixh,e
	  ld e,1
	  exx
	  ld iy,loop
	literal
	  ldi
	loop
	  GETBIT
	  jr nc,literal
	  exx
	  ld h,d
	  ld l,e
	getlen
	  GETBITEXX
	  jr nc,.lenok
	.lus
	  GETBITEXX
	  adc hl,hl
	  ret c
	  GETBITEXX
	  jr nc,.lenok
	  GETBITEXX
	  adc hl,hl
	  ret c
	  GETBITEXX
	  jp c,.lus
	.lenok
	  inc hl
	  exx
	  ld c,(hl)
	  inc hl
	  ld b,0
	  bit 7,c
	  jp z,offsok
	  jp ix

	mode6
	  GETBIT
	  rl b
	mode5
	  GETBIT
	  rl b
	mode4
	  GETBIT
	  rl b
	mode3
	  GETBIT
	  rl b
	mode2
	  GETBIT
	  rl b
	  GETBIT
	  jr nc,offsok
	  or a
	  inc b
	  res 7,c
	offsok
	  inc bc
	  push hl
	  exx
	  push hl
	  exx
	  ld l,e
	  ld h,d
	  sbc hl,bc
	  pop bc
	  ldir
	  pop hl
	  jp iy

	getbit
	  ld a,(hl)
	  inc hl
	  rla
	  ret

	getbitexx
	  exx
	  ld a,(hl)
	  inc hl
	  exx
	  rla
	  ret

	modes
	  word offsok
	  word mode2
	  word mode3
	  word mode4
	  word mode5
	  word mode6

	  endmodule

	;eof
	*/
}



