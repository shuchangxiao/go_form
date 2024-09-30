package utils

import (
	"golang.org/x/crypto/bcrypt"
	"unsafe"
)

func EncodePasswd(password string) (string, error) {
	fromPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return *(*string)(unsafe.Pointer(&fromPassword)), err
}
func EqualPasswd(hashpasswd string, passwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashpasswd), []byte(passwd))
	return err == nil
}
