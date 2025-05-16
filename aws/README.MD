# Uploader Backend API for AWS S3

Make sure that you install go compiler version more than 1.8 

If some path missing or need to run go get, run it.

## Prepare Go Workspace
Run this
Create Mod File
```bash
go mod init upload
```
Get all pre-requisites
```bash
go get "github.com/aws/aws-sdk-go-v2/aws"
go get "github.com/aws/aws-sdk-go-v2/config"
go get "github.com/aws/aws-sdk-go-v2/credentials"
go get "github.com/aws/aws-sdk-go-v2/service/s3"
go get "github.com/gin-gonic/gin"
go get "github.com/google/uuid"
go get "github.com/joho/godotenv"
go get "github.com/gin-contrib/cors"
```

## Setup and Get S3 Bucket
To Do

Create S3 Bucket
Get region / bucket name / accesskey / secretkey 

### Ensure that
1. Your access key have least privileges (GetObjects / PutObjects)
2. Allow Public to access S3 

## Setup Secret File
```text
APP_ADDRESS={listen address , default 127.0.0.1}
APP_PORT={listen port, default 5000}
PROVIDER=aws
APP_ALLOW_ORIGIN={web origin}
OSS_ACCESS_KEY={accesskey}
OSS_SECRET_KEY={secretkey}
OSS_REGION={region}
OSS_ENDPOINT=s3.{region}.amazonaws.com
OSS_PUBLIC_ENDPOINT=s3.{region}.amazonaws.com
OSS_BUCKET={bucket name}
OSS_PROTOCOL=https
```


## Compilation for test

once finished test, go live by final compile

```bash
go run main.go
```

App will run at provided address:port


## Binding with NGINX
```bash
   location /ip {
        if ($http_origin ~* "^https?://([a-z0-9.-]+\.)?domain\.(tld)$") {
            add_header Access-Control-Allow-Origin $http_origin;
            add_header Access-Control-Allow-Credentials true;
            add_header Content-Type application/json;
            return 200 '{"ip":"$remote_addr"}';
        }
        return 403;
    }
    location / {
                proxy_pass http://127.0.0.1:5000;
                proxy_set_header Upgrade $http_upgrade;
                proxy_set_header Connection 'upgrade';
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Client-IP $http_x_real_ip;
                error_page  404 /404.html;
                error_page 403 /403.html;
                error_page 444 /444.html;
                error_page 413 /413.html;
                error_page 500 502 503 504 /50x.html;
    }
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

Looking for hybrid provider in single binary

## License

[MIT](https://choosealicense.com/licenses/mit/)
