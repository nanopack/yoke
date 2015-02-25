package main

// import(
// 	"bufio"
// 	"errors"
// 	"fmt"
// 	"os"
// 	"strings"
// )

// //
// func ConfigureHBAConf() error {

// 	//
// 	entry := fmt.Sprintf(`
// ## additional configuration added by Pagoda Box ##
// host    replication     postgres        %s            trust
// `, "<otherguys ip>")

// 	//
// 	fi, err := stat("/data/postgres.conf")

// 	return nil
// }

// //
// func ConfigurePGConf() error {

// 	//
// 	entry := `
// ## additional configuration added by Pagoda Box ##
// wal_level 				= hot_standby #
// archive_mode 			= on 					#
// archive_command 	= 'exit 0' 		#
// max_wal_senders 	= 10 					#
// wal_keep_segments = 5000     		# 80 GB required on pg_xlog
// hot_standby 			= on 					#
// `
// 	// master only
// 	addendum := `
// synchronous_standby_names = slave
// `
// 	//
// 	fi, err := stat("/data/pg_hba.conf")

// 	// parse config file
// 	opts, err := parseFile("")
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// //
// func createRecovery() error {

// 	//
// 	entry := fmt.Sprintf(`
// ## additional configuration added by Pagoda Box ##
// standby_mode = on
// primary_conninfo = 'host=%s port=%s application_name=slave'
// restore_command = 'exit 0'
// `, "<otherguys ip>", "<other guys postgres port>")

// 	//
// 	fi, err := stat("/data/recovery.conf")

// 	return nil
// }

// //
// func destroyRecovery() error {
// 	//
// 	fi, err := stat("/data/recovery.conf")

// 	// err := os.Remove("/data/recovery.conf")
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	return nil
// }

// //
// func stat(f string) (os.FileInfo, error) {
// 	fi, err := os.Stat(f)
// 	if err != nil {
// 		log.Fatal("[pg_config.readFile]", err)
// 		return nil, err
// 	}

// 	return fi, nil
// }

// // parseFile will parse a config file, returning a 'opts' map of the resulting
// // config options.
// func parseFile(file string) (map[string]string, error) {

// 	// attempt to open file
// 	f, err := os.Open(file)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer f.Close()

// 	opts := make(map[string]string)
// 	scanner := bufio.NewScanner(f)
// 	readLine := 1

// 	// Read line by line, sending lines to parseLine
// 	for scanner.Scan() {
// 		if err := parseLine(scanner.Text(), opts); err != nil {
// 			log.Error("[pg_config] Error reading line: %v\n", readLine)
// 			return nil, err
// 		}

// 		readLine++
// 	}

// 	return opts, err
// }

// // parseLine reads each line of the config file, extracting a key/value pair to
// // insert into an 'opts' map.
// func parseLine(line string, m map[string]string) error {

// 	// handle instances where we just want to skip the line and move on
// 	switch {

// 	// skip empty lines
// 	case len(line) <= 0:
// 		return nil

// 	// skip commented lines
// 	case strings.HasPrefix(line, "#"):
// 		return nil
// 	}

// 	// extract key/value pair
// 	fields := strings.Fields(line)

// 	// ensure expected length of 2
// 	if len(fields) != 2 {
// 		return errors.New("Incorrect format. Expecting 'key value', received: " + line)
// 	}

// 	// insert key/value pair into map
// 	m[fields[0]] = fields[1]

// 	return nil
// }