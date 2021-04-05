// icsv_test.go
package icsv

import (
	"bytes"
	// "encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func aTest0(t *testing.T) {
	t.Errorf("os.args=%#v", os.Args)
	// t.Fail()
}

func test2(t *testing.T, txt1 string, sr *Reader, expect [][]string) {
	rtn, err := sr.ReadAll()
	if err != nil {
		t.Errorf("err=%#v", err)
	}
	for i, rtnRow := range rtn {
		for j, rtnCell := range rtnRow {
			if rtnCell != expect[i][j] {
				t.Errorf("txt=%#v", txt1)
				t.Errorf("expect=%#v", expect)
				t.Errorf("actural    =%#v", rtn)
				break
			}
		}
	}

}

func test1(t *testing.T, txt1 string, sr *Reader, expect [][]string) {
	rtn, err := sr.ReadAll()
	if err != nil {
		t.Errorf("err=%#v", err)
	}

	// fmt.Printf("%#v\n", txt1)
	// fmt.Printf("%#v\n", rtn)
	if fmt.Sprintf("%#v", rtn) != fmt.Sprintf("%#v", expect) {
		t.Errorf("txt=%#v", txt1)
		t.Errorf("expect =%#v", expect)
		t.Errorf("actural=%#v", rtn)
	}
}

func Test1(t *testing.T) {
	tests := map[string][][]string{
		/*
		*/
		"你好 \t": [][]string{[]string{"你好 \t"}},
		"a": [][]string{[]string{"a"}},
		`\\`: [][]string{[]string{`\\`}},
		`"\`: [][]string{[]string{`"\`}},
		"a\n": [][]string{[]string{"a"}},
		",": [][]string{[]string{"",""}},
		",\n": [][]string{[]string{"",""}},
		"a,": [][]string{[]string{"a", ""}},
		"a,\n": [][]string{[]string{"a", ""}},
		"a,b": [][]string{[]string{"a", "b"}},
		`a,"b\`: [][]string{[]string{`a`, `"b\`}},
		"a,b\n": [][]string{[]string{"a", "b"}},
		"a,b\n1": [][]string{[]string{"a", "b"},[]string{"1"}},

		// skip empty line
		"":         [][]string(nil),
		"\n":         [][]string(nil),

		// bom
		"\xEF\xBB\xBFno BOM": [][]string{[]string{"no BOM"}},

		/*
		*/
	 }

	for txt1, expect := range tests {
		r := NewReader(strings.NewReader(txt1))
		r.Comma = ','
		r.Terminator = '\n'
		test1(t, txt1, r, expect)
	}
}


func Test2(t *testing.T) {
	tests := map[string][][]string{
		/*
		*/
		"你好 \t": [][]string{[]string{"你好"}},
		"a": [][]string{[]string{"a"}},
		"a\n": [][]string{[]string{"a"}},
		",": [][]string{[]string{"",""}},
		",\n": [][]string{[]string{"",""}},
		"a,": [][]string{[]string{"a", ""}},
		"a,\n": [][]string{[]string{"a", ""}},
		"a,b": [][]string{[]string{"a", "b"}},
		"a,b\n": [][]string{[]string{"a", "b"}},
		"a,b\n1": [][]string{[]string{"a", "b"},[]string{"1"}},
		" a b\t\t\t": [][]string{[]string{"a b"}},

		// skip empty line
		"":         [][]string(nil),
		" ":        [][]string(nil),
		" \t ":     [][]string(nil),
		"\n":       [][]string(nil),
		"\n ":      [][]string(nil),
		"\n\n\n":   [][]string(nil),
		" \n \n\n": [][]string(nil),
		" \n \n# comment\n": [][]string(nil),
		" \n \n # comment\n": [][]string(nil),
		" \n \n\n # comment": [][]string(nil),

		// bom
		"\xEF\xBB\xBFno BOM": [][]string{[]string{"no BOM"}},

		// Test Comment
		"a #comment": [][]string{[]string{"a"}},
		`"#not comment"`: [][]string{[]string{`#not comment`}},

		// Test CharMapping
		// 口 is mapped space via CharMapping
		"口": [][]string{[]string{""}},
		`",口"`: [][]string{[]string{","}},

		// Test quotation marks and escape
		`"a"`: [][]string{[]string{`a`}},
		`" "" "`: [][]string{[]string{`"`}},
		`" "", "`: [][]string{[]string{`",`}},
		`" \" , "`: [][]string{[]string{`" ,`}},
		`" a    b "`: [][]string{[]string{`a    b`}},
		`" a    \, "`: [][]string{[]string{`a    ,`}},
		`" \"   \, "`: [][]string{[]string{`"   ,`}},
		` \"   \, `: [][]string{[]string{`"   ,`}},
		/*
		*/

		// Error cases
		// unclosed quote upto end of file, or upto max cell size
		`" `: [][]string(nil), // error to be raised
		`"a`: [][]string{[]string{`a`}},  // error to be raised
		// `"a"b"c`: [][]string{[]string{`a`}},  // error to be raised
		// first quote is not at beginning of the cell
		`a,b"`: [][]string{[]string{`a`,`b`}},  // error to be raised
		/*
		*/
	 }

	for txt1, expect := range tests {
		r := NewReader(strings.NewReader(txt1))

		r.Comma = ','
		r.Quote = '"'
		r.Escape = '\\'
		r.Comment =  '#'
		r.Terminator = '\n'
		// r.AroundTrim = "\n\r \t"
		r.LeadingTrim = "\n\r \t"
		r.TrailingTrim = "\n\r \t"
		r.CharMapping ="口 "

		test1(t, txt1, r, expect)
	}
}



func aTest2(t *testing.T) {
	// copy one io stream to two file
	txt1 := "\nHello-\n"
	sr := strings.NewReader(txt1)
	var buf bytes.Buffer
	tr := io.TeeReader(sr, &buf)
	// io.Copy(os.Stderr, sr)
	io.Copy(os.Stdout, tr)
}

func aTest9(t *testing.T) {
	var tests = []struct {
		in  int
		out int
	}{
		{1, 1},
		{2, 2},
		{3, 4},
	}
	for i, tt := range tests {
		if tt.in != tt.out {
			t.Errorf("test %d: %#v != %#v", i, tt.in, tt.out)
			// t.Fail()
		}
	}
}
