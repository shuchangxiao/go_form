package config

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"gopkg.in/gomail.v2"
	"log"
	"veripTest/global"
	"veripTest/model"
)

func InitRabbitMQ() {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", Cf.Rabbitmq.Username, Cf.Rabbitmq.Password, Cf.Rabbitmq.Host, Cf.Rabbitmq.Port)
	conn, err := amqp.Dial(url)
	failOnError(err, "连接RabbitMQ失败")
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	sendEmailQueue, err := ch.QueueDeclare(
		"sendEmail", // 队列名称
		false,       // 是否持久化
		false,       // 是否自动删除
		false,       // 是否独立
		false, nil,
	)
	failOnError(err, "Failed to create queue")
	global.Channel = ch
	global.SendEmailRoutineKey = sendEmailQueue.Name
	//body := "Hello World!"
	//err = ch.Publish(
	//	"",     // exchange
	//	sendEmailQueue.Name, // routing key
	//	false,  // mandatory
	//	false,  // immediate
	//	amqp.Publishing{
	//		ContentType: "text/plain",
	//		Body:        []byte(body),
	//	})
	//failOnError(err, "Failed to publish a message")
	msgs, err := ch.Consume(
		sendEmailQueue.Name, // queue
		"",                  // consumer
		true,                // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	failOnError(err, "Failed to register a consumer")
	// 申明一个goroutine,一遍程序始终监听
	go func() {
		for msg := range msgs {
			var sendEmail model.SendEmail
			err2 := json.Unmarshal(msg.Body, &sendEmail)
			if err2 != nil {
				failOnError(err2, "Failed to convert to sendEmail")
			}
			d := gomail.NewDialer(
				Cf.Email.Host,
				Cf.Email.Port,
				Cf.Email.Username,
				Cf.Email.Password,
			)
			m := gomail.NewMessage()
			m.SetHeader("From", Cf.Email.Username) // 发件人
			// m.SetHeader("From", "alias"+"<"+userName+">") // 增加发件人别名

			m.SetHeader("To", sendEmail.Email) // 收件人，可以多个收件人，但必须使用相同的 SMTP 连接
			m.SetHeader("Subject", "邮箱验证码")
			switch sendEmail.Status {
			case 1:
				m.SetBody("text/html", fmt.Sprintf("注册邮箱，您邮箱的验证码为: %d", sendEmail.Code))
			case 2:
				m.SetBody("text/html", fmt.Sprintf("忘记密码，您邮箱的验证码为: %d", sendEmail.Code))
			case 3:
				m.SetBody("text/html", fmt.Sprintf("修改密码，您邮箱的验证码为: %d", sendEmail.Code))
			default:
				log.Printf("出现未知status,%v", msg.Body)
				continue
			}
			d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
			err3 := d.DialAndSend(m)
			if err3 != nil {
				failOnError(err3, "Failed to email failure")
			}
			log.Printf("Received a message: %s", msg.Body)
		}
	}()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
