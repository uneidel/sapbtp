# SAP BTP

This is exported sample code for using golang with SAP BTP. This repo will be extended. 

The Following Functions are available at present:

  - helper 
    - flatten-json    
  - sap 
    - create-device-template 
    - create-device
  - iotcockpit 
    -  get-devices         
    - get-all-gateways
  - fakelogger 
    - send


**sap create-device-template**

This function takes a sample json file parses it and autogenerate device template.
Supported datatypes "integer" "boolean", "string","date"(RFC3339), "double

For all actions you need to provide a config.yml (fixed name)

    # IoT Cockpit
    Iotcockpit:
        IoTServiceCFAPIURL: "https://xxxxxx-58c3-414f-b854-b4237a4bfc9a.eu10.cp.iot.sap/c6d4df73-58c3-414f-b854-b4237a4bfc9a"
        Username: "root"
        Password: "XXXXXXXX"
        TenantId: "XXXXXXX"

    # Leonardo Credentials
    Leonardo:
        AuthUrl: "https://XXXXXXXX.authentication.eu10.hana.ondemand.com"
        ClientId: "sb-XXXXXXXX-4f53-4273-ae3e-f918347f632b!b11938|iotae_service!b5"
        ClientSecret: "XXXXXXXXX/hx+3/VRh/fC/ZbePWgw="
        TenantId: "xx.xx"



## WARRANTIES OR CONDITIONS

Unless required by applicable law or agreed to in writing, 
software distributed under the License is distributed on an "AS IS" BASIS, 
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.