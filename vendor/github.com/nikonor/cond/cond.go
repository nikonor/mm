package cond

import (
	"strconv"
	"strings"
)

// OK - функция проверки логического условия
//		Входящие параметры
//			in -  логическая строка
//			data - словарь переменных
//		Выходные переменные
//			bool - условие выполнено
//			err - ошибка
func OK(in string, data map[string]string) (bool, error) {

	in = strings.TrimSpace(in)

	if err := checkCond(in); err != nil {
		return false, err
	}

	in = fillFromMap(in, data)
	if strings.Contains(in, "$$") {
		return false, ErrIncompleteData
	}

	runeIn := []rune(in)
	cmdArray, max, _ := setIndexes(runeIn)

	// for m := 1; m <= max; m++ {
	// бежим по условиям снизу вверх
	for m := max; m >= 1; m-- {

		var (
			b, e  int
			flag  bool
			tasks []task
		)

		// ищем условия заданного уровня
		for idx, d := range cmdArray {
			if d == m {
				if !flag {
					b = idx
					flag = true
				} else {
					e = idx
					flag = false
					tasks = append(tasks, task{
						b: b,
						e: e,
						s: getCondition(runeIn, b, e),
					})
				}
			}
		}

		// TODO: делаем замену
		for _, t := range tasks {
			c := task2cond(t.s)
			if c == nil || c.iAction == 0 {
				return false, ErrWrongCommand
			}
			if c.iAction != CmdNot && len(c.s) == 0 {
				return false, ErrParseCond
			}
			p := "TRUE"
			if !c.do() {
				p = "FALSE"
			}
			runeIn = put(t.b, t.e+1, runeIn, []rune(p))
		}

	}

	if strings.TrimSpace(string(runeIn)) == F {
		return false, nil
	}

	return true, nil

}

func fillFromMap(in string, m map[string]string) string {
	for k, v := range m {
		if strings.Contains(v, " ") || len(v) == 0 {
			v = `"` + v + `"`
		}
		in = strings.Replace(in, "$$"+k+"$$", v, -1)
	}
	return in
}

func put(b, e int, sourse, sub []rune) []rune {
	sourceLen := len(sourse)
	subLen := len(sub)
	if b+subLen > sourceLen {
		return sourse
	}

	if e > sourceLen {
		e = sourceLen
	}

	count := 0
	for i := b; i < b+subLen; i++ {
		sourse[i] = sub[count]
		count++
	}
	for i := b + subLen; i < e; i++ {
		sourse[i] = ' '
	}
	return sourse
}

func task2cond(s []rune) *cond {
	var (
		c     cond
		ok    bool
		count int
	)
	ss := strings.TrimSpace(string(s[1 : len(s)-1]))
	sss := strings.Split(ss, " ") // rSpaces.Split(ss, -1)

	if len(sss) < 2 {
		return nil
	}

	c.sAction = sss[0]
	if c.iAction, ok = getIdxCmd(c.sAction); !ok {
		return nil

	}

	c.f, count = getString(sss[1:])
	if c.iAction != CmdNot {
		c.s, _ = getString(sss[1+count:])
	}

	return &c
}

func getString(ss []string) (string, int) {

	if len(ss) == 0 {
		return "", 0
	}
	if !strings.HasPrefix(ss[0], `"`) {
		for _, e := range ss {
			if len(e) > 0 {
				return e, 1
			}
		}
		return "", 0
	}

	var ret []string
	if ss[0] == `""` {
		ret = append(ret, ss[0])
	} else {
		ret = append(ret, ss[0][1:])
		var i int
		for i = 1; i < len(ss); i++ {
			if strings.HasSuffix(ss[i], `"`) && !strings.HasSuffix(ss[i], `\"`) {
				ret = append(ret, ss[i][:len(ss[i])-1])
				break
			}
			ret = append(ret, ss[i])
		}
	}

	return strings.Join(ret, " "), len(ret)
}

func getIdxCmd(s string) (int, bool) {
	switch s {
	case EQ:
		return CmdEQ, true

	case NE:
		return CmdNE, true

	case GT:
		return CmdGT, true

	case LT:
		return CmdLT, true

	case GTE:
		return CmdGTE, true

	case LTE:
		return CmdLTE, true

	case AND:
		return CmdAND, true

	case OR:
		return CmdOR, true

	case NOT:
		return CmdNot, true

	default:
		return 0, false
	}
}

func getCondition(in []rune, b int, e int) []rune {
	return in[b : e+1]
}

func setIndexes(in []rune) ([]int, int, error) {
	l := len(in)

	var (
		cmdArray  = make([]int, l)
		max, val  = 1, 1
		quote     = -1
		quoteFLag = false
	)
	for i := 0; i < l; i++ {
		switch {
		case in[i] == Open && !quoteFLag:
			cmdArray[i] = val
			max = setMax(max, val)
			val++
		case in[i] == Close && !quoteFLag:
			if quoteFLag {
				return nil, max, ErrParseCond
			}
			val--
			cmdArray[i] = val
		case in[i] == Quote:
			if !quoteFLag {
				cmdArray[i] = quote
				quoteFLag = true
			} else {
				cmdArray[i] = quote
				quote--
				quoteFLag = false
			}
		}
	}

	return cmdArray, max, nil
}

func setMax(max, cur int) int {
	if max >= cur {
		return max
	}
	return cur
}

func checkCond(in string) error {
	if len(in) < 3 ||
		(in[0] != '(' || in[len(in)-1] != ')') ||
		!strings.Contains(in, " ") {
		return ErrParseCond
	}

	if !checkBrackets(in) {
		return ErrParseCond
	}

	return nil
}

func checkBrackets(in string) bool {
	count := 0
	for i := 0; i < len(in); i++ {
		switch in[i] {
		case '(':
			count++
		case ')':
			count--
		}
		if count < 0 {
			return false
		}
	}

	return count == 0
}

// func countChar(in string, ch rune) int {
// 	ii := []rune(in)
// 	count := 0
// 	for _, i := range ii {
// 		if i == ch {
// 			count++
// 		}
// 	}
// 	return count
// }

func (c *cond) do() bool {
	switch c.iAction {
	case CmdEQ:
		return c.f == c.s
	case CmdNE:
		return c.f != c.s
	case CmdGT:
		return c.doDig()
	case CmdLT:
		return c.doDig()
	case CmdGTE:
		return c.doDig()
	case CmdLTE:
		return c.doDig()
	case CmdAND:
		return c.f == T && c.s == T
	case CmdOR:
		return c.f == T || c.s == T
	case CmdNot:
		return !(c.f == T)
	}
	return false
}

func (c *cond) doDig() bool {
	f, err := strconv.ParseFloat(c.f, 32)
	if err != nil {
		return false
	}
	s, err := strconv.ParseFloat(c.s, 32)
	if err != nil {
		return false
	}
	switch c.iAction {
	case CmdGT:
		return f > s
	case CmdLT:
		return f < s
	case CmdGTE:
		return f >= s
	case CmdLTE:
		return f <= s
	}
	return false
}
