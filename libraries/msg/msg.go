package msg

import (
	"GoLang/libraries/util"
	"github.com/kataras/iris"
)

const (
	CHANNEL_ADMIN_MESSAGE string = "admin_message"
	SEND_SYSTEM_MESSAGE   string = "SendSystemMessage"
	CHANNEL_PAYMENT       string = "payment"
)

func SystemMessage(mess string, users ...[]iris.Map) string {
	//users := []iris.Map{}
	//users = append(users, iris.Map{"user_id": "phuocnh", "server_id": 0})
	//users = append(users, iris.Map{"user_id": "phuocnh", "server_id": 1})

	if len(users) == 0 {
		data := iris.Map{
			"cmd": SEND_SYSTEM_MESSAGE,
			"data": iris.Map{
				"content":   mess,
				"list_user": "",
			},
		}
		return util.JsonEndCode(data)
	} else {
		data := iris.Map{
			"cmd": SEND_SYSTEM_MESSAGE,
			"data": iris.Map{
				"content":   mess,
				"list_user": users[0],
			},
		}
		return util.JsonEndCode(data)
	}
}
