package handler

type CreateDeviceCmd struct {
	Version string
	Ipv6    string
}

func (cmd *CreateDeviceCmd) Run() error {
	/*
		iotcockpit, leonardo := GetConfig()
		alternateId := strings.Replace(cmd.Ipv6, ":", "", -1)
		sensoralternateId := fmt.Sprintf("Sensor_%s", cmd.Ipv6)
		deviceName := fmt.Sprintf("DEVICE_%s", cmd.Ipv6)
		ctxID := fmt.Sprintf("%d", time.Now().UnixNano())

		var thingTemplate []string = []string{cmd.Version}

			// consider to move alternateId to uniqueId
			log.Printf("%s: Create Leonardo Thing with Template: %v", ctxID, thingTemplate)
			thing, err := leonardo.CreateThing(ctxID,
				cmd.Ipv6,
				alternateId,
				"Device"+alternateId,
				"v12",
				leonardo.Objectgroupid,
				"",
				thingTemplate)
			if err != nil {
				return fmt.Errorf("%s: cannot create thing: %s", ctxID, err)
			}
			log.Printf("%s: Thing with Id: %v created.", ctxID, thing.ID)

				log.Printf("%s: Create IoTCockpit Device", ctxID)
				deviceId, err := iotcockpit.CreateNewDevice(cmd.Ipv6, leonardo.GatewayId)
				if err != nil {
					log.Printf("%s: cannot create new device. Rollback: ", ctxID)
					err2 := leonardo.DeleteThing(ctxID, thing.ID)
					if err2 != nil {
						log.Printf("%s: cannot not delete thing with id: %s. Error: %s", ctxID, thing.ID, err2)
					}
					return err
				}

					log.Printf("%s: DeviceId: %v", ctxID, deviceId)
					log.Printf("%s: Create IoTCockpit Sensor", ctxID)
					sensorID, err := iotcockpit.CreateNewSensor(ipv6, deviceId, leonardo.SensortypeId)
					if err != nil {
						log.Printf("%s: cannot create IotCockpit Sensor. Rollback: %v", ctxID, err)
						err2 := iotcockpit.DeleteDevice(deviceId)
						if err2 != nil {
							log.Printf("%s: cannot delete device with id: %s. Error: %s", ctxID, deviceId, err2)
						}
						err2 = leonardo.DeleteThing(ctxID, thing.ID)
						if err2 != nil {
							log.Printf("%s: cannot delete thing with id: %s. Error: %s", ctxID, thing.ID, err2)
						}
						return err
					}
					log.Printf("[INFO] %s: SensorID: %v", ctxID, sensorID)

					err = iotcockpit.GetSingleSensor(sensorID)
					if err != nil {
						log.Printf("%s: sensor not yet created", ctxID)
					} else {
						log.Printf("%s: sensor responsive", ctxID)
					}

					//Create Assignment: Thing to Sensor
					log.Printf("%s: Create Assignment", ctxID)
					//CreateAssignment sometimes responses with "sensor not found (404)" even after get single sensor was already successfull
					var assignmentID, etag string
					err = nil
					for i := 0; i < 5; i++ {
						assignmentID, etag, err = leonardo.CreateAssignment(ctxID, thing.ID, sensorID, l.AssignmentMappingID)
						if err != nil {
							log.Printf("%s: cannot create assignment on try: %d, %v", ctxID, i, err)
							time.Sleep(500 * time.Millisecond)
						} else {
							break
						}
					}
					if err != nil {
						log.Error().Msgf("%s: cannot create assignment: %v, Rollback!", ctxID, err)
						err2 := l.IotCockpit.DeleteSensor(sensorID)
						if err2 != nil {
							log.Printf("%s: cannot delete sensor: %s", ctxID, err2)
						}
						err2 = l.IotCockpit.DeleteDevice(deviceId)
						if err2 != nil {
							log.Printf("%s: cannot delete device with id: %s. Error: %s", ctxID, deviceId, err2)
						}
						err2 = l.Leonardo.DeleteThing(ctxID, thing.ID)
						if err2 != nil {
							log.Printf("%s: cannot delete thing with id: %s. Error: %s", ctxID, thing.ID, err2)
						}
						return false, err
					}
					log.Info().Msgf("%s: Successfully created assignment: %s", ctxID, assignmentID)

					// Create Database Entries
					// Empty Loggerdata // foreign key constraints
					log.Info().Msgf("%s: Insert Empty Logger Data", ctxID)
					loggerDataId, err := database.InsertEmptyLoggerData()
					if err != nil {
						log.Error().Msgf("%s: cannot insert empty logger data. Rollback", ctxID)
						err2 := l.Leonardo.DeleteAssignment(ctxID, assignmentID, etag)
						if err2 != nil {
							log.Printf("%s: cannot delete assignment: %s", ctxID, err2)
						}
						err2 = l.IotCockpit.DeleteSensor(sensorID)
						if err2 != nil {
							log.Printf("%s: cannot delete sensor: %s", ctxID, err2)
						}
						err2 = l.IotCockpit.DeleteDevice(deviceId)
						if err2 != nil {
							log.Printf("%s: cannot delete device with id: %s. Error: %s", ctxID, deviceId, err2)
						}
						err2 = l.Leonardo.DeleteThing(ctxID, thing.ID)
						if err2 != nil {
							log.Printf("%s: cannot delete thing with id: %s. Error: %s", ctxID, thing.ID, err2)
						}
						return false, err
					}
					log.Info().Msgf("%s: Empty LoggerData ID: %v created.", ctxID, loggerDataId)
					// Empty loggersettings // foreign key constraints

					log.Info().Msgf("%s: Creating Logger in database", ctxID)
					// Create Loggerentry
					err = database.InsertLogger(uniqueId, ipv6, loggerDataId, clientgroup, version, -1, freeText, thing.ID, loggerKey, activated, imsi, iccid)
					if err != nil {
						log.Error().Msgf("%s: cannot insert logger in database. Rollback: %s", ctxID, err)
						//database.DeleteLoggerSettings()
						err2 := database.DeleteLoggerData(loggerDataId)
						if err2 != nil {
							log.Printf("%s: cannot delete logger data: %s", ctxID, err2)
						}
						sensoralternateId := fmt.Sprintf("Sensor_%s", ipv6)
						sensorId, err2 := l.IotCockpit.GetSensorIdbyAlternateId(sensoralternateId)
						if err2 != nil {
							log.Error().Msgf("%s: cannot get sensorId by alternate ID: %s", ctxID, err2)
						}
						err2 = l.IotCockpit.DeleteSensor(sensorId)
						if err2 != nil {
							log.Printf("%s: cannot delete sensor: %s", ctxID, err2)
						}
						err2 = l.IotCockpit.DeleteDevice(deviceId)
						if err2 != nil {
							log.Printf("%s: cannot delete device with id: %s. Error: %s", ctxID, deviceId, err2)
						}
						err2 = l.Leonardo.DeleteThing(ctxID, thing.ID)
						if err2 != nil {
							log.Printf("%s: cannot delete thing with id: %s. Error: %s", ctxID, thing.ID, err2)
						}
						return false, err
					}
					log.Info().Msgf("%s: logger with UniqueID: %v created.", ctxID, uniqueId)
					log.Info().Msgf("%s: Inserting Default Logger Settings for IPV6: %v", ctxID, ipv6)
					err = database.InsertDefaultLoggerSettings(ctxID, ipv6)
					if err != nil {
						log.Error().Msgf("%s: cannot insert Default LoggerSettings. Rollback.", ctxID)
						err2 := database.DeleteLoggerData(loggerDataId)
						if err2 != nil {
							log.Printf("%s: cannot delete logger data: %s", ctxID, err2)
						}
						sensoralternateId := fmt.Sprintf("Sensor_%s", ipv6)
						sensorId, err2 := l.IotCockpit.GetSensorIdbyAlternateId(sensoralternateId)
						if err2 != nil {
							log.Printf("%s: cannot get sensorID by alternate ID: %s", ctxID, err2)
						}
						err2 = l.IotCockpit.DeleteSensor(sensorId)
						if err2 != nil {
							log.Printf("%s: cannot delete sensor: %s", ctxID, err2)
						}
						err2 = l.IotCockpit.DeleteDevice(deviceId)
						if err2 != nil {
							log.Printf("%s: cannot delete device with id: %s. Error: %s", ctxID, deviceId, err2)
						}
						err2 = l.Leonardo.DeleteThing(ctxID, thing.ID)
						if err2 != nil {
							log.Printf("%s: cannot delete thing with id: %s. Error: %s", ctxID, thing.ID, err2)
						}
						return false, err

					}
					log.Info().Msgf("%s: Customer Default loggerSettings created.", ctxID)
	*/
	return nil
}
