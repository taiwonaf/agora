package main

import (
	"os"
	"log"
	"fmt"
	"strconv"
	"time"

	"github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	"github.com/gin-gonic/gin"
)

var appID, appCertificate string

func main() {
	appIDEnv, appIDExists := os.LookupEnv("APP_ID")
	appCertEnv, appCertExists := os.LookupEnv("APP_CERTIFICATE")

	if !appIDExists || !appCertExists {
		log.Println("Fatal error: Env not properly configured, check APP_ID and APP_CERTIFICATE ")
	} else {
		appID = appIDEnv
		appCertificate = appCertEnv
	}
	api := gin.Default()

	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	api.GET("rtc/:channelName/:role/:tokenType/:uid", getRtcToken)
	api.Run(":9090")
}
func getRtcToken(c *gin.Context) {
	
	channelName, tokenType, uid, role, expireTimeStamp, err := parseRtcParams(c)

	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{
			"message": "Error generating RTC tokem: " + err.Error(),
			"status": 400,
		})
		return
	}
	//generate rtc token
	rtcToken, tokenErr := generateRtcToken(channelName, uid, tokenType, role, expireTimeStamp)
	// return the token in JSON response
	if tokenErr {
		log.fatal(tokenErr)
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"error": "Error Generating RTC Token" + tokenErr.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"rtcToken": rtcToken,
		})
	}
}
func parseRtcParams(c *gin.Context) (channelName, tokenType, uidStr string, role rtctokenbuilder.Role, expireTimeStamp uint32, err error) {
	// get param values
	channelName = c.Param("channelName")
	roleStr := c.Param("role")
	tokenType = c.Param("tokenType")
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry", "3000")

	if roleStr == "publisher" {
		role = rtctokenbuilder.RolePublisher
	} else if roleStr == "subscriber" {
		role = rtctokenbuilder.RoleSubscriber
	} else {
		c.JSON(400, gin.H{
			"Error:": "role is invalid",
		})
	}

	expireTime64, parseErr := strenv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		err = fmt.Errorf("Failed to parse expireTime: %s, causing error: %s", expireTime, parseErr)
	}

	expireTimeInSeconds := uint32(expireTime64)
	currentTimeStamp := unit32(time.Now().UTC().unix())
	expireTimestamp = currentTimeStamp + expireTimeInSeconds

	return channelName, tokenType, uidStr, role, expireTimeStamp, err
}

func generateRtcToken(channelName, uidStr, tokenType string, role rtctokenbuilder.Role, expireTime uint32, err error) {
	// check token type
	if tokenType == "userAccount" {
		rtcToken , err = rtctokenbuilder.BuildTokenEithUserAccount(appID, appCertificate, channelName, uidStr, role, expireTimeStamp)
		return rtcToken, err
	} else if tokenType == "uid" {
		uid64, parseErr := strconv.ParseUint(uidStr, 10, 64)
		if parseErr != nil {
			err = fmt.Errorf("Failed to parse uidStr: %s, to uint causing error: %s ", uidStr, parseErr)
			return "", err
		}

		uid := uint32(uid64)
		rtcToken, err = rtctokenbuilder.BuildTokenWithUid(appID, appCertificate, channelName, uid, role, expireTimeStamp)
		return rtcToken , err
	} else {
		err = fmt.Errorf("Failed to generate RTC token for unknown tokentype: %s", tokenType)
		log.Println(err)
		return "", err
	}
}