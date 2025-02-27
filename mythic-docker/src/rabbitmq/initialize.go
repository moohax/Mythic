package rabbitmq

import (
	"fmt"
	"github.com/its-a-feature/Mythic/database"
	databaseStructs "github.com/its-a-feature/Mythic/database/structs"
	"github.com/its-a-feature/Mythic/grpc"
	"sync"
	"time"

	"github.com/its-a-feature/Mythic/logging"
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueHandler func(amqp.Delivery)
type RPCQueueHandler func(amqp.Delivery) interface{}

type RPCQueueStruct struct {
	Exchange   string
	Queue      string
	RoutingKey string
	Handler    RPCQueueHandler
}
type DirectQueueStruct struct {
	Exchange   string
	Queue      string
	RoutingKey string
	Handler    QueueHandler
}

type rabbitMQConnection struct {
	conn             *amqp.Connection
	mutex            sync.RWMutex
	addListenerMutex sync.RWMutex
	RPCQueues        []RPCQueueStruct
	DirectQueues     []DirectQueueStruct
}

var RabbitMQConnection rabbitMQConnection

func (r *rabbitMQConnection) AddRPCQueue(input RPCQueueStruct) {
	r.addListenerMutex.Lock()
	r.RPCQueues = append(r.RPCQueues, input)
	r.addListenerMutex.Unlock()
}
func (r *rabbitMQConnection) AddDirectQueue(input DirectQueueStruct) {
	r.addListenerMutex.Lock()
	r.DirectQueues = append(r.DirectQueues, input)
	r.addListenerMutex.Unlock()
}
func (r *rabbitMQConnection) startListeners() {
	exclusiveQueue := true
	for _, rpcQueue := range r.RPCQueues {
		go RabbitMQConnection.ReceiveFromRPCQueue(
			rpcQueue.Exchange,
			rpcQueue.Queue,
			rpcQueue.RoutingKey,
			rpcQueue.Handler,
			exclusiveQueue)
	}
	for _, directQueue := range r.DirectQueues {
		go RabbitMQConnection.ReceiveFromMythicDirectExchange(
			directQueue.Exchange,
			directQueue.Queue,
			directQueue.RoutingKey,
			directQueue.Handler,
			exclusiveQueue)
	}
	go checkContainerStatus()
}

func Initialize() {

	for {
		if _, err := RabbitMQConnection.GetConnection(); err == nil {
			// periodically check to make sure containers are online
			RabbitMQConnection.startListeners()
			// initialize the callback graph
			callbackGraph.Initialize()
			// re-spin up listening ports for proxies
			proxyPorts.Initialize()
			// start tracking tasks wait to be fetched
			submittedTasksAwaitingFetching.Initialize()
			grpc.Initialize()
			go func() {
				// wait 20s for things to stabilize a bit, then send a startup message
				time.Sleep(time.Second * 30)
				go emitStartupMessages()
			}()
			logging.LogInfo("RabbitMQ Initialized")
			return
		}
		logging.LogInfo("Waiting for RabbitMQ...")
	}
}

func emitStartupMessages() {
	operations := []databaseStructs.Operation{}
	if err := database.DB.Select(&operations, `SELECT "name", webhook, id, channel
		FROM operation WHERE complete=false AND deleted=false`); err != nil {
		logging.LogError(err, "Failed to fetch operations, so sending a generic one to everybody")
		RabbitMQConnection.EmitWebhookMessage(WebhookMessage{
			OperationID:      0,
			OperationName:    "",
			OperationWebhook: "",
			OperationChannel: "",
			OperatorUsername: "Mythic",
			Action:           WEBHOOK_TYPE_NEW_STARTUP,
			Data: map[string]interface{}{
				"startup_message": "Mythic Online!",
			},
		})
	} else {
		for _, op := range operations {
			RabbitMQConnection.EmitWebhookMessage(WebhookMessage{
				OperationID:      op.ID,
				OperationName:    op.Name,
				OperationWebhook: op.Webhook,
				OperationChannel: op.Channel,
				OperatorUsername: "Mythic",
				Action:           WEBHOOK_TYPE_NEW_STARTUP,
				Data: map[string]interface{}{
					"startup_message": fmt.Sprintf("Mythic Online for operation %s!", op.Name),
				},
			})
		}
	}

}
