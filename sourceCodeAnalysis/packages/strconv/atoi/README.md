#Atoi
字符串转整形.

### Atoi
Atoi函数是对外提供的接口函数.具体实现由ParseInt函数实现。

```
func Atoi(s string) (int, error) {
	const fnAtoi = "Atoi"

	// 字符串长度
	sLen := len(s)

	// int类型小整数通过快速模式运算
	if intSize == 32 && (0 < sLen && sLen < 10) ||
		intSize == 64 && (0 < sLen && sLen < 19) {
	
		s0 := s

		// 如果有符号表示正负,从第二位开始运算
		if s[0] == '-' || s[0] == '+' {
			s = s[1:]
		}

		// 字符串转整形运算
		n := 0
		for _, ch := range []byte(s) {
			ch -= '0'
			n = n*10 + int(ch)   // 按10进制运算
		}

		// 如果负数,转换完成之后,加上"-"号
		if s0[0] == '-' {
			n = -n
		}
		return n, nil
	}

	// int类型大整数,通过普通模式运算
	i64, err := ParseInt(s, 10, 0)
	if nerr, ok := err.(*NumError); ok {
		nerr.Func = fnAtoi
	}
	return int(i64), err
}
```


### ParseInt
ParseInt函数是Atoi函数的具体实现。该函数也对外提供类型转换。

```
// ParseInt参数:
// s - 待转换字符串
// base - 转换进制
// bitsize - 参数结果必须正确的整数类型,eg: int(0), int8(8), int16(16), int32(32), int64(64)
func ParseInt(s string, base int, bitSize int) (i int64, err error) {
	const fnParseInt = "ParseInt"


	// Pick off leading sign.
	s0 := s

	// 默认非负数
	neg := false

	// 如果要转换的整数类型,存在符号标示,从字符串第二位开始运算
	if s[0] == '+' {
		s = s[1:]
	} else if s[0] == '-' {
		// 负数
		neg = true
		s = s[1:]
	}

	// 处理无符号整数
	un, err = ParseUint(s, base, bitSize)
	if err != nil && err.(*NumError).Err != ErrRange {
		err.(*NumError).Func = fnParseInt
		err.(*NumError).Num = s0
		return 0, err
	}

	// 转成int64类型
	n := int64(un)
	if neg {
		n = -n
	}
	return n, nil
}
```