package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"veripTest/config"
	"veripTest/constant"
	"veripTest/global"
)

type Weather struct {
	Location map[string]interface{}
	Now      map[string]interface{}
	Hourly   []interface{}
}

func GetHeFengWeather(ctx *gin.Context) {
	var input struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	}
	weather := Weather{}
	if !global.InitPredicate(ctx, &input) {
		return
	}
	id, err := getForLocationId(&weather, input.Latitude, input.Longitude)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "获取天气失败",
		})
		ctx.Abort()
		return
	}
	cacheWeather, err := global.Redis.Get(constant.WEATHERCACHE + id).Result()
	if err == nil {
		err := json.Unmarshal([]byte(cacheWeather), &weather)
		if err != nil {
			log.Printf("从redis中获取天气信息出现错误：%v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"data":    nil,
				"message": "内部错误，请联系管理员",
			})
			ctx.Abort()
			return
		} else {
			ctx.JSON(http.StatusOK, gin.H{
				"data":    weather,
				"message": nil,
			})
			ctx.Abort()
			return
		}
	}
	err = fetchWeatherFromAPI(&weather, id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "获取天气失败",
		})
		ctx.Abort()
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":    weather,
		"message": nil,
	})
	marshal, err := json.Marshal(&weather)
	if err != nil {
		global.FailOnErr(ctx)
		log.Printf("转换成byte[]出现错误：%v", err)
		return
	}
	err = global.Redis.Set(constant.WEATHERCACHE+id, marshal, 1*time.Hour).Err()
	if err != nil {
		global.FailOnErr(ctx)
		log.Printf("redis中存储键值：%v", err)
		return
	}
	ctx.Abort()
	return
}
func getFromAPI(url string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("连接和风天气api时出现错误：%v", err)
		return nil, err
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("获取body时出现错误：%v", err)
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		log.Printf("关闭body连接时出现错误: %v", err)
		return nil, err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(bodyText, &data); err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return nil, err
	}
	return data, nil
}
func fetchWeatherFromAPI(weather *Weather, id string) error {
	urlNow := fmt.Sprintf("https://devapi.qweather.com/v7/weather/now?location=%s&key=%s", id, config.Cf.Weather.Key)
	data, err := getFromAPI(urlNow)
	if err != nil {
		return err
	}
	now, ok := data["now"].(map[string]interface{})
	if !ok {
		log.Printf("获取现在天气时出错:%v", data["now"])
		return errors.New("获取现在天气时出错")
	}
	//marshal, err := json.Marshal(now)
	//if err != nil {
	//	log.Printf("now转换成json时出现错误:%v", err)
	//	return errors.New("now转换成json时出现错误")
	//}
	weather.Now = now
	urlNext := fmt.Sprintf("https://devapi.qweather.com/v7/weather/24h?location=%s&key=%s", id, config.Cf.Weather.Key)
	data, err = getFromAPI(urlNext)
	if err != nil {
		return err
	}
	hourly, ok := data["hourly"].([]interface{})
	if !ok {
		log.Printf("获取24小时天气时出错:%v", data["hourly"])
		return errors.New("获取24小时天气时出错")
	}
	if len(hourly) > 0 {
		hourly = hourly[:5]
	}
	//marshal, err := json.Marshal(hourly)
	//if err != nil {
	//	log.Printf("hourly转换成json时出现错误:%v", err)
	//	return errors.New("hourly转换成json时出现错误")
	//}
	weather.Hourly = hourly
	return nil
}

func getForLocationId(weather *Weather, longitude float64, latitude float64) (id string, err error) {
	url := fmt.Sprintf("https://geoapi.qweather.com/v2/city/lookup?location=%f,%f&key=%s", latitude, longitude, config.Cf.Weather.Key)
	data, err := getFromAPI(url)
	if err != nil {
		return "", err
	}
	location, ok := data["location"].([]interface{})
	if !ok {
		log.Println("Location is not an array")
		return "", errors.New("location is not an array")
	}

	if len(location) > 0 {
		firstLocation, ok := location[0].(map[string]interface{})
		if !ok {
			log.Println("First location is not a map")
			return "", errors.New("first location is not a map")
		}
		//loc, err := json.Marshal(firstLocation)
		//if err != nil {
		//	log.Println("Error unmarshalling JSON:", err)
		//	return "", err
		//}
		weather.Location = firstLocation
		id, ok := firstLocation["id"].(string)
		if !ok {
			log.Println("City id is not a string")
			return "", errors.New("city id is not a string")
		} else {
			return id, nil
		}
	}
	return "", errors.New("获取不到数据")
}
