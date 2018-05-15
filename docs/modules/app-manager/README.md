## QuickStart
KubeSphere uses same app-manager module from [OpenPitrix](https://github/openpitrix/openpitrix), which is another open source project initiated by QingCloud. 
For testing and development purpose, follow steps below to setup app-manager service locally:
* Make sure git and docker runtime is installed in your local environment  
* Clone the OpenPitrix project to your local environment: 
  ```console
  git clone https://github.com/openpitrix/openpitrix.git
  ```  
* Get into openpitrix directory, run commands below:  
  ```console
  cd openpitrix
  make build
  make compose-up-app
  ```  

## Test app-manager

Visit http://127.0.0.1:9100/swagger-ui in browser, and try it online, or test app-manager api service via command line:

```shell
$ curl http://localhost:9100/v1/apps
{"total_items":0,"total_pages":0,"page_size":10,"current_page":1}