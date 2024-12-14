# Развертывание

## Подготовока окружения

Требования к окружению:

> [!important] ubuntu 22.04/ubuntu 24.04/debian 12/
> [!important] установленный docker и docker compose

```bash
apt update --yes
apt-get install docker.io docker-compose-v2 --yes
```

> [!note] если `docker-compose-v2` не найден - установите `docker-compose` и далее во всех вызовах `docker compose` используйте `docker-compose`

## Запуск приложения

```bash
git clone https://git.codenrock.com/5hm3l/jamel
```

```
Cloning into 'jamel'...
Username for 'https://git.codenrock.com': cnrprod1731062288-user-94510
Password for 'https://cnrprod1731062288-user-94510@git.codenrock.com':
warning: redirecting to https://git.codenrock.com/5hm3l/jamel.git/
remote: Enumerating objects: 10446, done.
remote: Counting objects: 100% (476/476), done.
remote: Compressing objects: 100% (435/435), done.
remote: Total 10446 (delta 186), reused 0 (delta 0), pack-reused 9970
Receiving objects: 100% (10446/10446), 50.64 MiB | 600.00 KiB/s, done.
Resolving deltas: 100% (2378/2378), done.
Updating files: 100% (8257/8257), done.
```

> [!info] данные для подключения доступны на [странице задачи](https://codenrock.com/contests/sovkombank-securehack/#/tasks/2206/6292)

```bash
cd jamel/
```

```bash
docker compose up -d
```

```
[+] Running 16/16
 ✔ jamel-server 1 layers [⣿]      0B/0B      Pulled                                                                                        86.7s
   ✔ 77b4f45193cd Pull complete                                                                                                            86.1s
 ✔ jamel-client 3 layers [⣿⣿⣿]      0B/0B      Pulled                                                                                     117.1s
   ✔ 38a8310d387e Already exists                                                                                                            0.0s
   ✔ 00e6563d6019 Pull complete                                                                                                            47.3s
   ✔ 6e2675045bcd Pull complete                                                                                                           116.3s
 ✔ minio 9 layers [⣿⣿⣿⣿⣿⣿⣿⣿⣿]      0B/0B      Pulled                                                                                      140.9s
   ✔ 2831c6e5194f Pull complete                                                                                                            32.8s
   ✔ f2f8f30a646a Pull complete                                                                                                             4.3s
   ✔ 3440aa9567dd Pull complete                                                                                                             0.2s
   ✔ 4414594dd510 Pull complete                                                                                                           139.7s
   ✔ c1cc85e2de65 Pull complete                                                                                                            52.5s
   ✔ d57a4fe62ee8 Pull complete                                                                                                            46.2s
   ✔ 48e0cffc0f68 Pull complete                                                                                                            47.1s
   ✔ 2b027acd57fe Pull complete                                                                                                            47.2s
   ✔ c1d0e26236f5 Pull complete                                                                                                            47.3s
[+] Running 4/4
 ✔ Container minio     Healthy                                                                                                              5.9s
 ✔ Container rabbitmq  Healthy                                                                                                             12.4s
 ✔ Container client    Started                                                                                                             12.6s
 ✔ Container server    Started
```

> [!important] Работа с системой возможна только после успешного обновления баз CVE

```bash
docker logs client
```

```
2024/12/14 16:40:37 loop started
2024/12/14 16:40:37 start update task
2024/12/14 16:45:39 updated finished
```

Дождитесь появления строки `updated finished` в логах - это значит, что базы обновлены и можно работать

Загрузите [бинарный файл для управления]() для соответствующей ос/архитектуры

```bash
jamel-admin_linux - linux amd64
jamel-admin_windows.exe - windows amd64
jamel-admin_darwin_arm - macos m1 и выше
jamel-admin_darwin_amd64 - macos intel
```

Запустите управляющий файл

```bash
./jamel-admin_linux

```

> [!note] при необходимости сделайте файл управляемым `chmod +x jamel-admin_linux` или добавьте файл в доверенные на своей macos

Взаимодействие админского(управляющего) файла с сервером осуществляется по сети, поэтому возможно подключение к удаленно развернутому серверу или подключение из другой сети. Для этого укажте адрес сервера в переменной окружения `SERVER=ip:port`

> [!info] стандартно управляющий файл подключается на 127.0.0.1:8443

```bash
SERVER=192.168.10.10:8443 ./jamel-admin_linux
```

или

```bash
export SERVER=192.168.10.10:8443
./jamel-admin_linux
```

# Использование

## Проверка образов `analyze`

## Работа с отчетами `report`

> [!todo] ЕГОР Опиши как использовать бинарь

# Описание инфраструктуры

> [!todo] ЕГОР Нарисуй схему и опиши как все работает
