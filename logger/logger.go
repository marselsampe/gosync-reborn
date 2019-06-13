// Copyright (c) 2018 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package logger

import (
	"github.com/sirupsen/logrus"
)

var testing = false

// SetLevel is to set logger level verbosity
func SetLevel(level string) (returnError error) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	return nil
}

// LogInfo is a helper function to make sure logging output conforms to logging format standard
func LogInfo(module string, error string, fmt string, args ...interface{}) {
	logrus.WithFields(logrus.Fields{"module": module, "resp": error}).
		Infof(fmt, args...)
}

// LogFatal is a helper function to make sure logging output conforms to logging format standard
func LogFatal(module string, error string, fmt string, args ...interface{}) {
	if testing {
		logrus.WithFields(logrus.Fields{"module": module, "resp": error}).
			Infof(fmt, args...)
	} else {
		logrus.WithFields(logrus.Fields{"module": module, "resp": error}).
			Fatalf(fmt, args...)
	}
}

// LogDebug is a helper function to make sure logging output conforms to logging format standard
func LogDebug(module string, error string, fmt string, args ...interface{}) {
	logrus.WithFields(logrus.Fields{"module": module, "resp": error}).
		Debugf(fmt, args...)
}

// LogError is a helper function to make sure logging output conforms to logging format standard
func LogError(module string, error string, fmt string, args ...interface{}) {
	logrus.WithFields(logrus.Fields{"module": module, "resp": error}).
		Errorf(fmt, args...)
}

// LogWarn is a helper function to make sure logging output conforms to logging format standard
func LogWarn(module string, error string, fmt string, args ...interface{}) {
	logrus.WithFields(logrus.Fields{"module": module, "resp": error}).
		Warnf(fmt, args...)
}

// Logf is a helper function to write directly to output without fields
func Logf(fmt string, args ...interface{}) {
	logrus.Printf(fmt, args...)
}

// Logln is a helper function to write directly to output without fields
func Logln(args ...interface{}) {
	logrus.Println(args...)
}
