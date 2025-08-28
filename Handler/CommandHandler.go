package Handler

import (
	"Gedis-Server/DB"
	"strconv"
	"strings"
)

type CommandHandler struct {
	db *DB.Database
}

func InitCommandHandler() *CommandHandler {
	return &CommandHandler{db: DB.GetDBInstance()}
}

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
	case ',':
		*index = *index + 1
		return ch.parseDoubles(resp, index)
	case '(':
		*index = *index + 1
		return ch.parseBigNums(resp, index)
	case '!':
		*index = *index + 1
		return ch.parseBulkErrors(resp, index)
	case '_':
		*index = *index + 1
		return []string{}
	case '=':
		*index = *index + 1
		return ch.parseVerbString(resp, index)
	default:
		return []string{}
	}
}

func (ch *CommandHandler) parseArray(str string, index *int) []string {
	var tokens []string
	ln := ""
	for str[*index] != '\r' {
		ln = ln + string(str[*index])
		*index = *index + 1
	}
	length, _ := strconv.Atoi(ln)
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
	ln := ""
	for str[*index] != '\r' {
		ln = ln + string(str[*index])
		*index = *index + 1
	}
	length, _ := strconv.Atoi(ln)
	*index = *index + 1
	tmp := *index
	tot := ""
	for *index < tmp+length {
		*index = *index + 1
		tot = tot + string(str[*index])
	}
	tokens = append(tokens, tot)
	return tokens
}

func (ch *CommandHandler) parseBulkErrors(str string, index *int) []string {
	return ch.parseBulkString(str, index)
}

func (ch *CommandHandler) parseVerbString(str string, index *int) []string {
	return ch.parseBulkString(str, index)
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

func (ch *CommandHandler) parseDoubles(str string, index *int) []string {
	return ch.parseSimpleError(str, index)
}

func (ch *CommandHandler) parseBigNums(str string, index *int) []string {
	return ch.parseSimpleError(str, index)
}

func (ch *CommandHandler) ExecCommand(str string) string {
	index := 0
	tokens := ch.parseResp(str, &index)
	if len(tokens) == 0 {
		return "-ERROR: Empty command\r\n"
	}
	command := tokens[0]
	command = strings.ToUpper(command)
	switch command {
	case "SET":
		return ch.db.Set(tokens)
	case "GET":
		return ch.db.Get(tokens)
	case "PING":
		return "+PONG\r\n"
	case "ECHO":
		return echoReply(tokens)
	case "FLUSHALL":
		return ch.db.FlushAll(tokens)
	case "KEYS":
		return ch.db.Keys(tokens)
	case "TYPE":
		return ch.db.Type(tokens)
	case "DEL":
		return ch.db.Del(tokens)
	case "UNLINK":
		return ch.db.Del(tokens)
	case "HSET":
		return ch.db.Hset(tokens)
	case "EXPIRE":
		return ch.db.Expire(tokens)
	case "RENAME":
		return ch.db.Rename(tokens)
	case "LLEN":
		return ch.db.Llen(tokens)
	default:
		return "-ERR UNKNOWN COMMAND\r\n"
	}
}

func echoReply(tokens []string) string {
	if len(tokens) < 2 {
		return "-ERR: Echo must have a message\r\n"
	}
	str := tokens[1]
	return "$" + strconv.Itoa(len(str)) + "\r\n" + str + "\r\n"
}
