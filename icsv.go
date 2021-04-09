package icsv

import (
	"bufio"
	"errors"
	"io"
)

type Reader struct {
	r *bufio.Reader

	// attributes in standard package
	Comma           rune // default ,
	FieldsPerRecord int
	Comment         rune // default #

	// extra attributes
	Quote      rune //  default "
	Escape     rune //  default \
	Terminator rune //  default \n

	// Trim white space or any char
	// for example char in `"'\n\r\t ,|;`
	AroundTrim   string // default "'\n\r\t\" ,|;"
	LeadingTrim  string // extra for leading trim
	TrailingTrim string // extra for trailing trim
	leadingTrim  map[rune]bool
	trailingTrim map[rune]bool

	CharMapping string        // "\n ,-"
	charMapping map[rune]rune // {'\n': ' ', ',':'-'}

	// internal variables
	beginFlag bool // for trimming BOM
	lastChar  rune

	cellNo   int
	recordNo int
	line     int // line number
	col      int // col number
	// initialized bool
	// columnCount int
	// lineNum int
	// lineTxt string

	// Future attributes
	// Replace [][]string  // [[`""`, `"`]]
	// Report string

	// if there is no new line at the end of line
	// 0: do nothing, 1: warning, 2: error
	// NewLineEOF uint8

	// BufferSize int // use for detecting csv config
	// Dialect string  // rcf, excel, mysql, postgres, informix, probe
	// Delimiter string  // , \t | ||
}

const BOM rune = 65279

var (
	CsvParsingError = errors.New("csv parsing error")
)

// Read till end of cell
// return: a cell, end of record or other error
func (r *Reader) readCell() (cellStr string, recordEnd uint8, err error) {
	readNo := 0     // char number in the read loop
	charNo := -1    // char number of a cell
	var ch rune     // char
	var sz int      // size of char
	var cell []rune // data cell

	// flags and cursors
	var comment bool
	var startQuot bool
	var endQuot bool
	var endQuotIdx int
	var escape bool

	charRemap := func(ch rune) rune {
		if val, ok := r.charMapping[ch]; ok {
			return val
		}
		return ch
	}

	addToCell := func() {
		charNo += 1
		cell = append(cell, charRemap(ch))
	}

Loop:
	for {
		ch, sz, err = r.r.ReadRune()
		readNo++
		r.col++

		// recordEnd=0: record is not ended
		// =1: when "...a\n..." // more lines to be read
		// =2: when "...a<EOF>" // This could be a broken file
		// =3: when "...a\n<EOF>"
		// =4: when "...a\n  <EOF>" // may be a broken file
		// =5: when "<EOF>"  // empty file
		if sz == 0 {
			recordEnd = 2
			if r.lastChar == '\n' {
				recordEnd = 3
			}
			if readNo > 1 && charNo == -1 {
				recordEnd = 4
			}
			if r.lastChar == 0 {
				recordEnd = 5
			}
			break
		} else {
			r.lastChar = ch
		}

		if ch == r.Terminator {
			r.line++
			r.col = 0
		}

		if !r.beginFlag {
			r.beginFlag = true
			r.trimMap()

			// skip BOM 0xEF,0xBB,0xBF
			if ch == BOM {
				continue
			}
		}

		if comment && ch != r.Terminator {
			continue
		}

		switch {
		case escape:
			escape = false
			// next escape tokens will tread as normal char
			// \"=>"  \,=>,
			addToCell()
			break

		// skip empty line
		case r.cellNo == 0 && charNo == -1 && ch == r.Terminator:
			break

		// if not in quotation string
		case !startQuot && ch == r.Comma: // end of cell if not in quoted string
			break Loop
		case !startQuot && ch == r.Terminator: // end of cell and record if not in quoted string
			recordEnd = 1
			break Loop
		case !startQuot && ch == r.Quote: // start quoted string
			startQuot = true
			break
		case !startQuot && ch == r.Comment: // start comment
			comment = true
			break

		case ch == r.Escape && !escape: // escape token
			escape = true
			break

		// trim leading space
		case charNo == -1 && r.leadingTrim[ch]:
			break

		// if in quotation string
		// this could be the 2nd quotation mark or the 4th
		case ch == r.Quote && startQuot && !endQuot:
			endQuotIdx = readNo
			endQuot = true
			break
		// this is just next to the 2nd quotation mark,
		// this is the end of quoted string and the end of the cell
		case ch == r.Comma && startQuot && endQuot && readNo == endQuotIdx+1:
			startQuot = false
			endQuot = false
			break Loop
		// same as above this is just next to the 2nd quotation mark,
		// "" double quotation marks means one in rcf csv standard
		case ch == r.Quote && startQuot && endQuot && readNo == endQuotIdx+1:
			endQuot = false
			addToCell()
			break
		default:
			addToCell()
		}
	}

	for i0 := charNo; i0 >= 0; i0-- {
		if r.trailingTrim[cell[i0]] {
			charNo = i0 - 1
		} else {
			break
		}
	}

	cellStr = string(cell[:charNo+1])
	return
}

// helper function
func stringToCharBool(rtn map[rune]bool, s string) {
	for _, c := range s {
		rtn[c] = true
	}
}

// helper function
func stringToCharMap(rtn map[rune]rune, str string) {
	var s []rune
	for _, c := range str {
		s = append(s, c)
	}

	l := len(s)
	for i := 0; i < l/2; i++ {
		rtn[s[i]] = s[i+1]
	}
}

// Initialize maps for trim cell string
func (r *Reader) trimMap() {
	r.charMapping = make(map[rune]rune)
	r.leadingTrim = make(map[rune]bool)
	r.trailingTrim = make(map[rune]bool)
	stringToCharMap(r.charMapping, r.CharMapping)
	stringToCharBool(r.leadingTrim, r.AroundTrim)
	stringToCharBool(r.trailingTrim, r.AroundTrim)
	stringToCharBool(r.leadingTrim, r.LeadingTrim)
	stringToCharBool(r.trailingTrim, r.TrailingTrim)
}

// Read till end of line or end of file
// return: a record, end of file or other error
func (r *Reader) Read() (rec []string, err error) {
	var cell string
	var recordEnd uint8
	r.cellNo = 0

	for {
		cell, recordEnd, err = r.readCell()
		if recordEnd > 2 {
			break
		}

		r.cellNo++
		rec = append(rec, cell)

		if recordEnd > 0 {
			break
		}
		if err != nil {
			break
		}
	}

	if r.cellNo > 0 {
		r.recordNo++
	}

	return
}

// Read All
// return: all records
func (r *Reader) ReadAll() (records [][]string, err error) {
	for {
		rec, err := r.Read()
		if rec != nil {
			records = append(records, rec)
		}
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
	}
	return
}

// NewReader returns a new Reader that reads from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:          bufio.NewReader(r),
		Terminator: '\n',
		Comma:      ',',
		// Quote: '"',
		// AroundTrim: "\n\r \t",
		// CharMapping: "\n \r \t ,.",
		// Escape: '\\',
		// Comment:  '#',
		// AroundTrim: "\n\r\t \",",
	}
}
