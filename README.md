# Менеджер паролей GophKeeper

### Выпускной проект курса Яндекс Практикум "Продвинутый Go-разработчик".

## Общие требования

GophKeeper представляет собой клиент-серверную систему, позволяющую пользователю надёжно и безопасно хранить логины, пароли, бинарные данные и прочую приватную информацию.

### Типы хранимой информации
* пары логин/пароль
* произвольные текстовые данные
* произвольные бинарные данные
* данные банковских карт

## Реализация
### Использованные технологии
* [Redis](https://github.com/redis/go-redis) - хранение шифрованных секретов и метаданных сервера
* [jwt](https://pkg.go.dev/github.com/golang-jwt/jwt/v4)   - аутентификация пользователей во время работы с сервером
* [grpc](https://protobuf.dev/reference/go/go-generated/)  - протокол взаимодействия клиент-сервер
* [cobra](https://github.com/spf13/cobra) - библиотека для cli интерфейса клиента

### Сервер

Сервер реализован на базе сгенерированного шаблона унарного grpc сервера. Для хранения технической информации и данных пользователей использована база Redis.

Регистрация новых пользователей и дальнейшая их аутентификация происходит через пару логин/пароль.

После аутентификации пользователю отправляется jwt токен с ограниченным временем жизни. Этот токен сохраняется клиентом в файл и используется в дальнейшем для запросов данных.

Так как система должна хранить и передавать данные безопасно для коммуникации используется шифрование tls протоколом. Сертификат tls генерится автоматически при каждом запуске сервера.

Данные пользователя храняться в зашифрованном виде индивидуальным ключом пользователя. Этот ключ также храниться в базе в зашифрованном виде мастер ключом сервера. Мастер ключ сервера передается при запуске сервера через флаг. Если при первом запуске сервера на пустой базе данных ключ не был предоставлен, то сервер сгенерирует его автоматически и отобразит в консольном выводе. При дальнейших запусках/перезапусках сервера на этой же базе необходимо предоставлять этот же ключ. В случае утери мастер ключа, база данных будет в зашифрованном виде и расшифровать ее будет не возможно.

Скачать сервер для linux можно [здесь](https://github.com/hrapovd1/gokeepas/tree/release/bin/linux).

#### Быстрый старт

```BASH
docker run --rm --name gokeepas-stor -d -p 6379:6379 redis:6-alpine redis-server

./keeppas-server -a 0.0.0.0:5000
```
```
	!!!! Not found server key, generate new: 

		'YXjGEdqZWeTZ255dLanqFV8s'	

	Please remember and provide it in the next run with this db !!!
{"level":"info","timestamp":"2023-05-19T07:27:11Z","caller":"server/main.go:89","msg":"server started"}
```

После чего можно подключаться к серверу на порт 5000/tcp.

### Клиент

Клиент реализован на базе библиотеки [cobra](https://github.com/spf13/cobra).

Для получения справки достаточно запустить клиент без параметров.

Скачать клиента для своей платформы можно [здесь](https://github.com/hrapovd1/gokeepas/tree/release/bin).

#### Подключение к серверу

Для успешного подключения к серверу необходимо использовать DNS имя или алиас из hosts файла, по ip адресу подключение не пройдет из-за ошибки сертификата.

```BASH
echo "SERVER_IP    keeppas" >> /etc/hosts
./keeppas -s keeppas:5000 signup -u USER_LOGIN -p USER_PASSWORD
```
```
login success
```

### Сборка

```BASH
mkdir -p bin/linux
mkdir -p bin/win
mkdir -p bin/mac
# build linux
go build -ldflags "-X 'github.com/hrapovd1/gokeepas/internal/cli.BuildTime=$(date +'%Y-%m-%d %H:%M')'" -o bin/linux/keeppas cmd/client/main.go
go build -o bin/linux/keeppas-server cmd/server/main.go
# build windows
GOOS=windows GOARCH=amd64 go build -ldflags "-X 'github.com/hrapovd1/gokeepas/internal/cli.BuildTime=$(date +'%Y-%m-%d %H:%M')'" -o bin/win/keeppas.exe cmd/client/main.go
# build mac amd64
GOOS=darwin GOARCH=amd64 go build -ldflags "-X 'github.com/hrapovd1/gokeepas/internal/cli.BuildTime=$(date +'%Y-%m-%d %H:%M')'" -o bin/mac/keeppas cmd/client/main.go
```
