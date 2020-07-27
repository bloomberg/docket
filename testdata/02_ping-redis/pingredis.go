// Copyright 2019 Bloomberg Finance L.P.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pingredis

import (
	"github.com/go-redis/redis/v7"
)

// PingRedis pings the Redis instance at redisAddr.
func PingRedis(redisAddr string) (string, error) {
	client := redis.NewClient(&redis.Options{Addr: redisAddr})

	return client.Ping().Result()
}
