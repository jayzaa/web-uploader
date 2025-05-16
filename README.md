# API Uploader for AWS S3 and AlibabaCloud OSS
## Warning
Try as your own responsibility

## Included
Example HTML file to use as upload form (dummy cookie for username, js for fetch visitor IP)

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

Looking for hybrid provider in single binary

## License

No License. However, if you are looking for cloud provider. Simply register by my referal link

[AlibabaCloud]https://www.alibabacloud.com/campaign/benefits?referral_code=A9ESHA


## Prerequisites
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
See Cloud Provider's folder
```

## Setup and Get S3/OSS Bucket
To Do

Create S3/OSS Bucket
Get region / bucket name / accesskey / secretkey 

### Ensure that
1. Your access key have least privileges (GetObjects / PutObjects)
2. Allow Public to access Storage

## Setup Secret File
```text
See Cloud Provider's folder
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
        ### View Visitor IP Address through nginx 
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


