package restful


func ResponsePack(errCode int64) map[string]interface{} {
	return map[string]interface{}{
		"Action":  "",
		"Result":  "",
		"Error":   errCode,
		"Desc":    ErrMap[errCode],
		"Version": "1.0.0",
	}
}


var ErrMap = map[int64]string{
	SUCCESS: "SUCCESS",
	PARA_ERROR:"PARAMETER ERROR",
	PARA_PARSE_ERROR:"PARA_PARSE_ERROR",
	QueryAllExcelFileName:"QueryAllExcelFileName",
}

const (
	SUCCESS = 1
	PARA_ERROR = 40000
	PARA_PARSE_ERROR = 40001
	QueryAllExcelFileName = 40002
)
