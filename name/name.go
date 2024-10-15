package name

import (
	"regexp"
	"strings"
)

func ValidateBucketName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}

	if !(isLowerLetterOrDigit(name[0]) && isLowerLetterOrDigit(name[len(name)-1])) {
		return false
	}

	re := regexp.MustCompile(`^[a-z0-9.-]+$`)
	if !re.MatchString(name) {
		return false
	}

	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "-") {
		return false
	}

	if strings.Contains(name, "./") || strings.Contains(name, "../") {
		return false
	}

	if strings.Contains(name, "..") || strings.Contains(name, "--") {
		return false
	}

	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if ipRegex.MatchString(name) {
		return false
	}

	return true
}

func isLowerLetterOrDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z')
}

// ValidateObjectKey проверяет, является ли имя объекта допустимым.
func ValidateObjectKey(key string) bool {
	// Проверяем длину имени объекта
	if len(key) < 1 || len(key) > 1024 {
		return false
	}

	// Проверяем, чтобы имя не содержало запрещенных символов
	re := regexp.MustCompile(`^[^<>:"/\\|?*\x00-\x1F]+$`)
	return re.MatchString(key)
}
