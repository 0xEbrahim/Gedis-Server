package main

type CommandHandler struct {
}

func initCommandHandler() *CommandHandler {
	return &CommandHandler{}
}

/**
*2\r\n$2ME\r\n$9ljsdhchjdchjdsc\r\n
 */

func (ch *CommandHandler) parseResp(resp string, index *int) []string {

	switch resp[*index] {
	case '*':
		*index = *index + 1
		return ch.parseArray(resp, index)
	case '+':
		*index = *index + 1
		return ch.parseSimpleString(resp, index)
	case '$':
		*index = *index + 1
		return ch.parseBulkString(resp, index)
	case '-':
		*index = *index + 1
		return ch.parseSimpleError(resp, index)
	case '#':
		*index = *index + 1
		return ch.parseBoolean(resp, index)
	default:
		return []string{}
	}
}

func (ch *CommandHandler) parseArray(str string, index *int) []string {
	var tokens []string
	length := str[*index] - '0'
	for length > 0 {
		*index = *index + 1
		if str[*index] == '\r' || str[*index] == '\n' {
			continue
		}
		tokens = append(tokens, ch.parseResp(str, index)...)
		length--
	}
	return tokens
}

func (ch *CommandHandler) parseSimpleString(str string, index *int) []string {
	var tokens []string
	tmp := ""
	for str[*index] != '\r' {
		tmp = tmp + string(str[*index])
		*index = *index + 1
	}
	tokens = append(tokens, tmp)
	return tokens
}

func (ch *CommandHandler) parseBulkString(str string, index *int) []string {
	var tokens []string
	length := str[*index] - '0'
	*index = *index + 2
	tmp := *index
	tot := ""
	for *index < tmp+int(length) {
		*index = *index + 1
		tot = tot + string(str[*index])
	}
	tokens = append(tokens, tot)
	return tokens
}

func (ch *CommandHandler) parseSimpleError(str string, index *int) []string {
	return ch.parseBulkString(str, index)
}

func (ch *CommandHandler) parseIntegers(str string, index *int) []string {
	return ch.parseSimpleString(str, index)
}

func (ch *CommandHandler) parseBoolean(str string, index *int) []string {
	return ch.parseSimpleError(str, index)
}
