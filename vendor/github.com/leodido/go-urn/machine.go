package urn

import (
	"fmt"
)

var (
	errPrefix         = "expecting the prefix to be the \"urn\" string (whatever case) [col %d]"
	errIdentifier     = "expecting the identifier to be string (1..31 alnum chars, also containing dashes but not at its start) [col %d]"
	errSpecificString = "expecting the specific string to be a string containing alnum, hex, or others ([()+,-.:=@;$_!*']) chars [col %d]"
	errNoUrnWithinID  = "expecting the identifier to not contain the \"urn\" reserved string [col %d]"
	errHex            = "expecting the specific string hex chars to be well-formed (%%alnum{2}) [col %d]"
	errParse          = "parsing error [col %d]"
)


const start int = 1
const first_final int = 44

const en_fail int = 46
const en_main int = 1


// Machine is the interface representing the FSM
type Machine interface {
	Error() error
	Parse(input []byte) (*URN, error)
}

type machine struct {
	data           []byte
	cs             int
	p, pe, eof, pb int
	err            error
	tolower        []int
}

// NewMachine creates a new FSM able to parse RFC 2141 strings.
func NewMachine() Machine {
	m := &machine{}

	return m
}

// Err returns the error that occurred on the last call to Parse.
//
// If the result is nil, then the line was parsed successfully.
func (m *machine) Error() error {
	return m.err
}

func (m *machine) text() []byte {
	return m.data[m.pb:m.p]
}

// Parse parses the input byte array as a RFC 2141 string.
func (m *machine) Parse(input []byte) (*URN, error) {
	m.data = input
	m.p = 0
	m.pb = 0
	m.pe = len(input)
	m.eof = len(input)
	m.err = nil
	m.tolower = []int{}
	output := &URN{}

	{
		m.cs = start
	}


	{
		if (m.p) == (m.pe) {
			goto _test_eof
		}
		switch m.cs {
		case 1:
			goto st_case_1
		case 0:
			goto st_case_0
		case 2:
			goto st_case_2
		case 3:
			goto st_case_3
		case 4:
			goto st_case_4
		case 5:
			goto st_case_5
		case 6:
			goto st_case_6
		case 7:
			goto st_case_7
		case 8:
			goto st_case_8
		case 9:
			goto st_case_9
		case 10:
			goto st_case_10
		case 11:
			goto st_case_11
		case 12:
			goto st_case_12
		case 13:
			goto st_case_13
		case 14:
			goto st_case_14
		case 15:
			goto st_case_15
		case 16:
			goto st_case_16
		case 17:
			goto st_case_17
		case 18:
			goto st_case_18
		case 19:
			goto st_case_19
		case 20:
			goto st_case_20
		case 21:
			goto st_case_21
		case 22:
			goto st_case_22
		case 23:
			goto st_case_23
		case 24:
			goto st_case_24
		case 25:
			goto st_case_25
		case 26:
			goto st_case_26
		case 27:
			goto st_case_27
		case 28:
			goto st_case_28
		case 29:
			goto st_case_29
		case 30:
			goto st_case_30
		case 31:
			goto st_case_31
		case 32:
			goto st_case_32
		case 33:
			goto st_case_33
		case 34:
			goto st_case_34
		case 35:
			goto st_case_35
		case 36:
			goto st_case_36
		case 37:
			goto st_case_37
		case 38:
			goto st_case_38
		case 44:
			goto st_case_44
		case 39:
			goto st_case_39
		case 40:
			goto st_case_40
		case 45:
			goto st_case_45
		case 41:
			goto st_case_41
		case 42:
			goto st_case_42
		case 43:
			goto st_case_43
		case 46:
			goto st_case_46
		}
		goto st_out
	st_case_1:
		switch (m.data)[(m.p)] {
		case 85:
			goto tr1
		case 117:
			goto tr1
		}
		goto tr0
	tr0:
		m.err = fmt.Errorf(errParse, m.p)
		(m.p)--

		{
			goto st46
		}

		goto st0
	tr3:
		m.err = fmt.Errorf(errPrefix, m.p)
		(m.p)--

		{
			goto st46
		}

		m.err = fmt.Errorf(errParse, m.p)
		(m.p)--

		{
			goto st46
		}

		goto st0
	tr6:
		m.err = fmt.Errorf(errIdentifier, m.p)
		(m.p)--

		{
			goto st46
		}

		m.err = fmt.Errorf(errParse, m.p)
		(m.p)--

		{
			goto st46
		}

		goto st0
	tr41:
		m.err = fmt.Errorf(errSpecificString, m.p)
		(m.p)--

		{
			goto st46
		}

		m.err = fmt.Errorf(errParse, m.p)
		(m.p)--

		{
			goto st46
		}

		goto st0
	tr44:
		m.err = fmt.Errorf(errHex, m.p)
		(m.p)--

		{
			goto st46
		}

		m.err = fmt.Errorf(errSpecificString, m.p)
		(m.p)--

		{
			goto st46
		}

		m.err = fmt.Errorf(errParse, m.p)
		(m.p)--

		{
			goto st46
		}

		goto st0
	tr50:
		m.err = fmt.Errorf(errPrefix, m.p)
		(m.p)--

		{
			goto st46
		}

		m.err = fmt.Errorf(errIdentifier, m.p)
		(m.p)--

		{
			goto st46
		}

		m.err = fmt.Errorf(errParse, m.p)
		(m.p)--

		{
			goto st46
		}

		goto st0
	tr52:
		m.err = fmt.Errorf(errNoUrnWithinID, m.p)
		(m.p)--

		{
			goto st46
		}

		m.err = fmt.Errorf(errIdentifier, m.p)
		(m.p)--

		{
			goto st46
		}

		m.err = fmt.Errorf(errParse, m.p)
		(m.p)--

		{
			goto st46
		}

		goto st0
	st_case_0:
	st0:
		m.cs = 0
		goto _out
	tr1:
		m.pb = m.p

		goto st2
	st2:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof2
		}
	st_case_2:
		switch (m.data)[(m.p)] {
		case 82:
			goto st3
		case 114:
			goto st3
		}
		goto tr0
	st3:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof3
		}
	st_case_3:
		switch (m.data)[(m.p)] {
		case 78:
			goto st4
		case 110:
			goto st4
		}
		goto tr3
	st4:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof4
		}
	st_case_4:
		if (m.data)[(m.p)] == 58 {
			goto tr5
		}
		goto tr0
	tr5:
		output.prefix = string(m.text())

		goto st5
	st5:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof5
		}
	st_case_5:
		switch (m.data)[(m.p)] {
		case 85:
			goto tr8
		case 117:
			goto tr8
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto tr7
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto tr7
			}
		default:
			goto tr7
		}
		goto tr6
	tr7:
		m.pb = m.p

		goto st6
	st6:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof6
		}
	st_case_6:
		switch (m.data)[(m.p)] {
		case 45:
			goto st7
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st7
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st7
			}
		default:
			goto st7
		}
		goto tr6
	st7:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof7
		}
	st_case_7:
		switch (m.data)[(m.p)] {
		case 45:
			goto st8
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st8
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st8
			}
		default:
			goto st8
		}
		goto tr6
	st8:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof8
		}
	st_case_8:
		switch (m.data)[(m.p)] {
		case 45:
			goto st9
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st9
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st9
			}
		default:
			goto st9
		}
		goto tr6
	st9:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof9
		}
	st_case_9:
		switch (m.data)[(m.p)] {
		case 45:
			goto st10
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st10
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st10
			}
		default:
			goto st10
		}
		goto tr6
	st10:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof10
		}
	st_case_10:
		switch (m.data)[(m.p)] {
		case 45:
			goto st11
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st11
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st11
			}
		default:
			goto st11
		}
		goto tr6
	st11:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof11
		}
	st_case_11:
		switch (m.data)[(m.p)] {
		case 45:
			goto st12
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st12
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st12
			}
		default:
			goto st12
		}
		goto tr6
	st12:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof12
		}
	st_case_12:
		switch (m.data)[(m.p)] {
		case 45:
			goto st13
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st13
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st13
			}
		default:
			goto st13
		}
		goto tr6
	st13:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof13
		}
	st_case_13:
		switch (m.data)[(m.p)] {
		case 45:
			goto st14
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st14
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st14
			}
		default:
			goto st14
		}
		goto tr6
	st14:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof14
		}
	st_case_14:
		switch (m.data)[(m.p)] {
		case 45:
			goto st15
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st15
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st15
			}
		default:
			goto st15
		}
		goto tr6
	st15:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof15
		}
	st_case_15:
		switch (m.data)[(m.p)] {
		case 45:
			goto st16
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st16
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st16
			}
		default:
			goto st16
		}
		goto tr6
	st16:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof16
		}
	st_case_16:
		switch (m.data)[(m.p)] {
		case 45:
			goto st17
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st17
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st17
			}
		default:
			goto st17
		}
		goto tr6
	st17:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof17
		}
	st_case_17:
		switch (m.data)[(m.p)] {
		case 45:
			goto st18
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st18
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st18
			}
		default:
			goto st18
		}
		goto tr6
	st18:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof18
		}
	st_case_18:
		switch (m.data)[(m.p)] {
		case 45:
			goto st19
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st19
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st19
			}
		default:
			goto st19
		}
		goto tr6
	st19:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof19
		}
	st_case_19:
		switch (m.data)[(m.p)] {
		case 45:
			goto st20
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st20
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st20
			}
		default:
			goto st20
		}
		goto tr6
	st20:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof20
		}
	st_case_20:
		switch (m.data)[(m.p)] {
		case 45:
			goto st21
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st21
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st21
			}
		default:
			goto st21
		}
		goto tr6
	st21:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof21
		}
	st_case_21:
		switch (m.data)[(m.p)] {
		case 45:
			goto st22
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st22
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st22
			}
		default:
			goto st22
		}
		goto tr6
	st22:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof22
		}
	st_case_22:
		switch (m.data)[(m.p)] {
		case 45:
			goto st23
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st23
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st23
			}
		default:
			goto st23
		}
		goto tr6
	st23:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof23
		}
	st_case_23:
		switch (m.data)[(m.p)] {
		case 45:
			goto st24
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st24
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st24
			}
		default:
			goto st24
		}
		goto tr6
	st24:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof24
		}
	st_case_24:
		switch (m.data)[(m.p)] {
		case 45:
			goto st25
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st25
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st25
			}
		default:
			goto st25
		}
		goto tr6
	st25:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof25
		}
	st_case_25:
		switch (m.data)[(m.p)] {
		case 45:
			goto st26
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st26
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st26
			}
		default:
			goto st26
		}
		goto tr6
	st26:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof26
		}
	st_case_26:
		switch (m.data)[(m.p)] {
		case 45:
			goto st27
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st27
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st27
			}
		default:
			goto st27
		}
		goto tr6
	st27:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof27
		}
	st_case_27:
		switch (m.data)[(m.p)] {
		case 45:
			goto st28
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st28
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st28
			}
		default:
			goto st28
		}
		goto tr6
	st28:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof28
		}
	st_case_28:
		switch (m.data)[(m.p)] {
		case 45:
			goto st29
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st29
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st29
			}
		default:
			goto st29
		}
		goto tr6
	st29:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof29
		}
	st_case_29:
		switch (m.data)[(m.p)] {
		case 45:
			goto st30
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st30
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st30
			}
		default:
			goto st30
		}
		goto tr6
	st30:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof30
		}
	st_case_30:
		switch (m.data)[(m.p)] {
		case 45:
			goto st31
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st31
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st31
			}
		default:
			goto st31
		}
		goto tr6
	st31:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof31
		}
	st_case_31:
		switch (m.data)[(m.p)] {
		case 45:
			goto st32
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st32
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st32
			}
		default:
			goto st32
		}
		goto tr6
	st32:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof32
		}
	st_case_32:
		switch (m.data)[(m.p)] {
		case 45:
			goto st33
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st33
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st33
			}
		default:
			goto st33
		}
		goto tr6
	st33:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof33
		}
	st_case_33:
		switch (m.data)[(m.p)] {
		case 45:
			goto st34
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st34
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st34
			}
		default:
			goto st34
		}
		goto tr6
	st34:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof34
		}
	st_case_34:
		switch (m.data)[(m.p)] {
		case 45:
			goto st35
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st35
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st35
			}
		default:
			goto st35
		}
		goto tr6
	st35:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof35
		}
	st_case_35:
		switch (m.data)[(m.p)] {
		case 45:
			goto st36
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st36
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st36
			}
		default:
			goto st36
		}
		goto tr6
	st36:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof36
		}
	st_case_36:
		switch (m.data)[(m.p)] {
		case 45:
			goto st37
		case 58:
			goto tr10
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st37
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st37
			}
		default:
			goto st37
		}
		goto tr6
	st37:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof37
		}
	st_case_37:
		if (m.data)[(m.p)] == 58 {
			goto tr10
		}
		goto tr6
	tr10:
		output.ID = string(m.text())

		goto st38
	st38:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof38
		}
	st_case_38:
		switch (m.data)[(m.p)] {
		case 33:
			goto tr42
		case 36:
			goto tr42
		case 37:
			goto tr43
		case 61:
			goto tr42
		case 95:
			goto tr42
		}
		switch {
		case (m.data)[(m.p)] < 48:
			if 39 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 46 {
				goto tr42
			}
		case (m.data)[(m.p)] > 59:
			switch {
			case (m.data)[(m.p)] > 90:
				if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
					goto tr42
				}
			case (m.data)[(m.p)] >= 64:
				goto tr42
			}
		default:
			goto tr42
		}
		goto tr41
	tr42:
		m.pb = m.p

		goto st44
	st44:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof44
		}
	st_case_44:
		switch (m.data)[(m.p)] {
		case 33:
			goto st44
		case 36:
			goto st44
		case 37:
			goto st39
		case 61:
			goto st44
		case 95:
			goto st44
		}
		switch {
		case (m.data)[(m.p)] < 48:
			if 39 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 46 {
				goto st44
			}
		case (m.data)[(m.p)] > 59:
			switch {
			case (m.data)[(m.p)] > 90:
				if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
					goto st44
				}
			case (m.data)[(m.p)] >= 64:
				goto st44
			}
		default:
			goto st44
		}
		goto tr41
	tr43:
		m.pb = m.p

		goto st39
	st39:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof39
		}
	st_case_39:
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st40
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st40
			}
		default:
			goto tr46
		}
		goto tr44
	tr46:
		m.tolower = append(m.tolower, m.p-m.pb)

		goto st40
	st40:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof40
		}
	st_case_40:
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st45
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st45
			}
		default:
			goto tr48
		}
		goto tr44
	tr48:
		m.tolower = append(m.tolower, m.p-m.pb)

		goto st45
	st45:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof45
		}
	st_case_45:
		switch (m.data)[(m.p)] {
		case 33:
			goto st44
		case 36:
			goto st44
		case 37:
			goto st39
		case 61:
			goto st44
		case 95:
			goto st44
		}
		switch {
		case (m.data)[(m.p)] < 48:
			if 39 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 46 {
				goto st44
			}
		case (m.data)[(m.p)] > 59:
			switch {
			case (m.data)[(m.p)] > 90:
				if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
					goto st44
				}
			case (m.data)[(m.p)] >= 64:
				goto st44
			}
		default:
			goto st44
		}
		goto tr44
	tr8:
		m.pb = m.p

		goto st41
	st41:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof41
		}
	st_case_41:
		switch (m.data)[(m.p)] {
		case 45:
			goto st7
		case 58:
			goto tr10
		case 82:
			goto st42
		case 114:
			goto st42
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st7
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st7
			}
		default:
			goto st7
		}
		goto tr6
	st42:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof42
		}
	st_case_42:
		switch (m.data)[(m.p)] {
		case 45:
			goto st8
		case 58:
			goto tr10
		case 78:
			goto st43
		case 110:
			goto st43
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st8
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st8
			}
		default:
			goto st8
		}
		goto tr50
	st43:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof43
		}
	st_case_43:
		if (m.data)[(m.p)] == 45 {
			goto st9
		}
		switch {
		case (m.data)[(m.p)] < 65:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st9
			}
		case (m.data)[(m.p)] > 90:
			if 97 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 122 {
				goto st9
			}
		default:
			goto st9
		}
		goto tr52
	st46:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof46
		}
	st_case_46:
		switch (m.data)[(m.p)] {
		case 10:
			goto st0
		case 13:
			goto st0
		}
		goto st46
	st_out:
	_test_eof2:
		m.cs = 2
		goto _test_eof
	_test_eof3:
		m.cs = 3
		goto _test_eof
	_test_eof4:
		m.cs = 4
		goto _test_eof
	_test_eof5:
		m.cs = 5
		goto _test_eof
	_test_eof6:
		m.cs = 6
		goto _test_eof
	_test_eof7:
		m.cs = 7
		goto _test_eof
	_test_eof8:
		m.cs = 8
		goto _test_eof
	_test_eof9:
		m.cs = 9
		goto _test_eof
	_test_eof10:
		m.cs = 10
		goto _test_eof
	_test_eof11:
		m.cs = 11
		goto _test_eof
	_test_eof12:
		m.cs = 12
		goto _test_eof
	_test_eof13:
		m.cs = 13
		goto _test_eof
	_test_eof14:
		m.cs = 14
		goto _test_eof
	_test_eof15:
		m.cs = 15
		goto _test_eof
	_test_eof16:
		m.cs = 16
		goto _test_eof
	_test_eof17:
		m.cs = 17
		goto _test_eof
	_test_eof18:
		m.cs = 18
		goto _test_eof
	_test_eof19:
		m.cs = 19
		goto _test_eof
	_test_eof20:
		m.cs = 20
		goto _test_eof
	_test_eof21:
		m.cs = 21
		goto _test_eof
	_test_eof22:
		m.cs = 22
		goto _test_eof
	_test_eof23:
		m.cs = 23
		goto _test_eof
	_test_eof24:
		m.cs = 24
		goto _test_eof
	_test_eof25:
		m.cs = 25
		goto _test_eof
	_test_eof26:
		m.cs = 26
		goto _test_eof
	_test_eof27:
		m.cs = 27
		goto _test_eof
	_test_eof28:
		m.cs = 28
		goto _test_eof
	_test_eof29:
		m.cs = 29
		goto _test_eof
	_test_eof30:
		m.cs = 30
		goto _test_eof
	_test_eof31:
		m.cs = 31
		goto _test_eof
	_test_eof32:
		m.cs = 32
		goto _test_eof
	_test_eof33:
		m.cs = 33
		goto _test_eof
	_test_eof34:
		m.cs = 34
		goto _test_eof
	_test_eof35:
		m.cs = 35
		goto _test_eof
	_test_eof36:
		m.cs = 36
		goto _test_eof
	_test_eof37:
		m.cs = 37
		goto _test_eof
	_test_eof38:
		m.cs = 38
		goto _test_eof
	_test_eof44:
		m.cs = 44
		goto _test_eof
	_test_eof39:
		m.cs = 39
		goto _test_eof
	_test_eof40:
		m.cs = 40
		goto _test_eof
	_test_eof45:
		m.cs = 45
		goto _test_eof
	_test_eof41:
		m.cs = 41
		goto _test_eof
	_test_eof42:
		m.cs = 42
		goto _test_eof
	_test_eof43:
		m.cs = 43
		goto _test_eof
	_test_eof46:
		m.cs = 46
		goto _test_eof

	_test_eof:
		{
		}
		if (m.p) == (m.eof) {
			switch m.cs {
			case 44, 45:
				raw := m.text()
				output.SS = string(raw)
				// Iterate upper letters lowering them
				for _, i := range m.tolower {
					raw[i] = raw[i] + 32
				}
				output.norm = string(raw)

			case 1, 2, 4:
				m.err = fmt.Errorf(errParse, m.p)
				(m.p)--

				{
					goto st46
				}

			case 3:
				m.err = fmt.Errorf(errPrefix, m.p)
				(m.p)--

				{
					goto st46
				}

				m.err = fmt.Errorf(errParse, m.p)
				(m.p)--

				{
					goto st46
				}

			case 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 41:
				m.err = fmt.Errorf(errIdentifier, m.p)
				(m.p)--

				{
					goto st46
				}

				m.err = fmt.Errorf(errParse, m.p)
				(m.p)--

				{
					goto st46
				}

			case 38:
				m.err = fmt.Errorf(errSpecificString, m.p)
				(m.p)--

				{
					goto st46
				}

				m.err = fmt.Errorf(errParse, m.p)
				(m.p)--

				{
					goto st46
				}

			case 42:
				m.err = fmt.Errorf(errPrefix, m.p)
				(m.p)--

				{
					goto st46
				}

				m.err = fmt.Errorf(errIdentifier, m.p)
				(m.p)--

				{
					goto st46
				}

				m.err = fmt.Errorf(errParse, m.p)
				(m.p)--

				{
					goto st46
				}

			case 43:
				m.err = fmt.Errorf(errNoUrnWithinID, m.p)
				(m.p)--

				{
					goto st46
				}

				m.err = fmt.Errorf(errIdentifier, m.p)
				(m.p)--

				{
					goto st46
				}

				m.err = fmt.Errorf(errParse, m.p)
				(m.p)--

				{
					goto st46
				}

			case 39, 40:
				m.err = fmt.Errorf(errHex, m.p)
				(m.p)--

				{
					goto st46
				}

				m.err = fmt.Errorf(errSpecificString, m.p)
				(m.p)--

				{
					goto st46
				}

				m.err = fmt.Errorf(errParse, m.p)
				(m.p)--

				{
					goto st46
				}

			}
		}

	_out:
		{
		}
	}

	if m.cs < first_final || m.cs == en_fail {
		return nil, m.err
	}

	return output, nil
}
