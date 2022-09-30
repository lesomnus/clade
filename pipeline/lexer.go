package pipeline

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Token int

const (
	TokenErr Token = iota
	TokenEOF

	TokenLeftParen  // '('
	TokenRightParen // ')'
	TokenPipe       // '|'

	TokenText
	TokenString // quoted string including quotes.
)

type Pos struct {
	Line   uint
	Column uint
}

type TokenItem struct {
	Pos   Pos
	Token Token
	Value string
}

type Lexer struct {
	pos Pos
	r   *bufio.Reader
}

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		pos: Pos{Line: 1, Column: 0},
		r:   bufio.NewReader(r),
	}
}

func (l *Lexer) readRune() (rune, TokenItem, error) {
	l.pos.Column++

	r, _, err := l.r.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return 0, TokenItem{Pos: l.pos, Token: TokenEOF, Value: ""}, nil
		}

		return 0, TokenItem{Pos: l.pos}, err
	}

	return r, TokenItem{}, nil
}

func (l *Lexer) unreadRune() {
	l.pos.Column--

	if err := l.r.UnreadRune(); err != nil {
		panic(err)
	}
}

func (l *Lexer) Lex() (TokenItem, error) {
	for {
		pos := l.pos

		r, t, err := l.readRune()
		if err != nil {
			return t, err
		}

		if t.Token == TokenEOF {
			return t, nil
		}

		switch r {
		case '\n':
			l.pos = Pos{l.pos.Line + 1, 0}
		case '(':
			return TokenItem{pos, TokenLeftParen, "("}, nil
		case ')':
			return TokenItem{pos, TokenRightParen, ")"}, nil
		case '|':
			return TokenItem{pos, TokenPipe, "|"}, nil

		case '`':
			l.unreadRune()
			return l.lexString()

		default:
			if unicode.IsSpace(r) {
				continue
			}

			if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("-+", r) {
				l.unreadRune()
				return l.lexText()
			}
		}
	}
}

func (l *Lexer) lexText() (TokenItem, error) {
	pos := l.pos

	var v string
	for {
		r, t, err := l.readRune()
		if err != nil {
			return t, err
		}

		if unicode.IsSpace(r) || strings.ContainsRune("()|`\n", r) {
			l.unreadRune()
			return TokenItem{pos, TokenText, v}, nil
		}

		if t.Token == TokenEOF {
			return TokenItem{pos, TokenText, v}, nil
		}

		v += string(r)
		continue
	}
}

func (l *Lexer) lexString() (TokenItem, error) {
	pos := l.pos

	var v string = "`"

	if r, t, err := l.readRune(); err != nil {
		return t, err
	} else if r != '`' {
		panic(fmt.Sprintf("string: expected ` but was %c", r))
	}

	for {
		r, t, err := l.readRune()
		if err != nil {
			return t, err
		}

		if t.Token == TokenEOF {
			return TokenItem{l.pos, TokenEOF, ""}, errors.New("unexpected EOF")
		}

		// End of string.
		if r == '`' {
			v += string(r)
			return TokenItem{pos, TokenString, v}, nil
		}

		if r == '\n' {
			return TokenItem{l.pos, TokenErr, string(r)}, errors.New("unexpected newline character")
		}

		v += string(r)
		continue
	}
}
