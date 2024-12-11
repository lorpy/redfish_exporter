# Server Exporter
### 请求示例

```text
http://localhost:9234/metrics?target=172.16.34.1
```

### 支持品牌

* HPE
* Dell
* Lenovo
* Huawei
* Inspur

### 编译方法

```sh
go build main.go
```

### 使用方法

```sh
usage:
    -dict: 原生的redfish似乎不支持采集物理机名（找了好久没找到），所以我做了这个映射关系，让exporter通过ip去拿到映射的物理机名字，格式见map.json
    -conf: 配置文件目录
    -prometheus(可选): 指定这个参数可以开启远程写模式
```

### 运行方法

```sh
./server_exporter -dict map.json -conf config.yml -prometheus http://1.1.1.1:9090
```

### 配置文件

```yaml
basic:
  port: 1230        #监听端口
  bindIp: 127.0.0.1 #监听ip
hosts:
  172.113.32.1:
    username: user
    password: pass
  172.113.32.2:
    username: user
    password: pass
metrics:
  system: true
  sensors: true
  power: true
  events: false
  storage: false
  memory: false
  network: false
```

### 指标

#### system

```text
idrac_system_power_on 1
idrac_system_health{status="OK"} 0
idrac_system_indicator_led_on{state="Lit"} 1
idrac_system_memory_size_bytes 137438953472
idrac_system_cpu_count{model="Intel(R) Xeon(R) Gold 6130 CPU @ 2.10GHz"} 2
idrac_system_bios_info{version="2.3.10"} 1
idrac_system_machine_info{manufacturer="Dell Inc.",model="PowerEdge C6420",serial="abc",sku="xyz"} 1
```

#### sensor

```text
idrac_sensors_temperature{id="0",name="Inlet Temp",units="celsius"} 19
idrac_sensors_fan_health{id="0",name="FAN1A",status="OK"} 0
idrac_sensors_fan_speed{id="0",name="FAN1A",units="rpm"} 7912
```

#### power

```text
idrac_power_supply_health{id="0",status="OK"} 0
idrac_power_supply_output_watts{id="0"} 74.5
idrac_power_supply_input_watts{id="0"} 89
idrac_power_supply_capacity_watts{id="0"} 750
idrac_power_supply_input_voltage{id="0"} 232
idrac_power_supply_efficiency_percent{id="0"} 91
```

```text
idrac_power_control_consumed_watts{id="0",name="System Power Control"} 166
idrac_power_control_capacity_watts{id="0",name="System Power Control"} 816
idrac_power_control_min_consumed_watts{id="0",name="System Power Control"} 165
idrac_power_control_max_consumed_watts{id="0",name="System Power Control"} 177
idrac_power_control_avg_consumed_watts{id="0",name="System Power Control"} 166
idrac_power_control_interval_in_minutes{id="0",name="System Power Control"} 1
```

#### sel

```text
idrac_events_log_entry{id="1",message="The process of installing an operating system or hypervisor is successfully completed",severity="OK"} 1631175352
```

#### storage

```text
idrac_drive_info{id="Disk.Direct.1-1:AHCI.Slot.5-1",manufacturer="MICRON",mediatype="SSD",model="MTFDDAV240TDU",name="SSD 1",protocol="SATA",serial="xyz",slot="1"} 1
idrac_drive_health{id="Disk.Direct.1-1:AHCI.Slot.5-1",status="OK"} 0
idrac_drive_capacity_bytes{id="Disk.Direct.1-1:AHCI.Slot.5-1"} 240057409536
idrac_drive_life_left_percent{id="Disk.Direct.1-1:AHCI.Slot.5-1"} 100
```

#### memory

```text
idrac_memory_module_info{ecc="MultiBitECC",id="DIMM.Socket.A2",manufacturer="Micron Technology",name="DIMM A2",rank="2",serial="xyz",type="DDR4"} 1
idrac_memory_module_health{id="DIMM.Socket.A2",status="OK"} 0
idrac_memory_module_capacity_bytes{id="DIMM.Socket.A2"} 34359738368
idrac_memory_module_speed_mhz{id="DIMM.Socket.A2"} 2400
```

#### network

```text
idrac_network_interface_health{id="NIC.Embedded.1",status="OK"} 0
idrac_network_port_health{id="NIC.Embedded.1-1",interface="NIC.Embedded.1",status="OK"} 0
idrac_network_port_link_up{id="NIC.Embedded.1-1",interface="NIC.Embedded.1",status="Up"} 1
idrac_network_port_speed_mbps{id="NIC.Embedded.1-1",interface="NIC.Embedded.1"} 1000
```
