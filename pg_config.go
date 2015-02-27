package main

import(
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

//
func configureHBAConf() error {

	//
  other, err := Whois(otherRole(myself()))
  if err != nil {
  	log.Warn("[pg_config.configureHBAConf] Unable to find another!\n%s\n", err)
  }

	//
	file := conf.DataDir+"pg_hba.conf"
	f, err := os.Open(file)
	if err != nil {
		log.Error("[pg_config.configureHBAConf] Failed to open '%s'!\n%s\n", file, err)
		return err
	}

	defer f.Close()

	//
	scanner := bufio.NewScanner(f)
	reFindConfigOption := regexp.MustCompile(`^\s*#?\s*(local|host)\s*(replication)`)
	readLine 	:= 1
	entry 		:= ""

	// Read file line by line
	for scanner.Scan() {

		// dont care about submatches, just if the string matches
		if reFindConfigOption.FindString(scanner.Text()) == "" {
			entry += fmt.Sprintf("%s\n", scanner.Text())
		}

		readLine++
	}

  //
	if other != nil {
		entry += fmt.Sprintf(`
#------------------------------------------------------------------------------
# PAGODA BOX
#------------------------------------------------------------------------------

# these configuration options have been removed from their standard location and
# placed here so that Pagoda Box could override them with the neccessary values
# to configure redundancy.

# IMPORTANT: these settings will always be overriden when the server boots. They
# are set dynamically and so should never change.

host    replication     postgres        %s/32            trust`, other.Ip)
	}

	//
	err = ioutil.WriteFile(file, []byte(entry), 0644)
  if err != nil {
  	log.Error("[pg_config.configureHBAConf] Failed to write to '%s'!\n%s\n", file, err)
  	return err
  }

	return nil
}

//
func configurePGConf(master bool) error {

	//
	file := conf.DataDir+"postgresql.conf"
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	defer f.Close()

	reFindConfigOption := regexp.MustCompile(`^\s*#?\s*(listen_addresses|port|wal_level|archive_mode|archive_command|max_wal_senders|wal_keep_segments|hot_standby)\s*=\s*`)
	scanner := bufio.NewScanner(f)
	readLine := 1
	entry := ""

	// Read file line by line
	for scanner.Scan() {

		submatch := reFindConfigOption.FindStringSubmatch(scanner.Text())

		//
		if submatch == nil {
			entry += fmt.Sprintf("%s\n", scanner.Text())
		}

		readLine++
	}

	//
	entry += fmt.Sprintf(`
#------------------------------------------------------------------------------
# PAGODA BOX
#------------------------------------------------------------------------------

# these configuration options have been removed from their standard location and
# placed here so that Pagoda Box could override them with the neccessary values
# to configure redundancy.

# IMPORTANT: these settings will always be overriden when the server boots. They
# are set dynamically and so should never change.

listen_addresses = 0.0.0.0        # what IP address(es) to listen on;
                                  # comma-separated list of addresses;
                                  # defaults to 'localhost'; use '*' for all
                                  # (change requires restart)
port = %d                     # (change requires restart)
wal_level = hot_standy            # minimal, archive, or hot_standby
                                  # (change requires restart)
archive_mode = on                 # allows archiving to be done
                                  # (change requires restart)
archive_command = 'exit 0'        # command to use to archive a logfile segment
                                  # placeholders: %p = path of file to archive
                                  #               %f = file name only
                                  # e.g. 'test ! -f /mnt/server/archivedir/%f && cp %p /mnt/server/archivedir/%f'
max_wal_senders = 10              # max number of walsender processes
                                  # (change requires restart)
wal_keep_segments = 5000          # in logfile segments, 16MB each; 0 disables
hot_standby = on                  # "on" allows queries during recovery
                                  # (change requires restart)`, conf.PGPort)

	//
	if master {
		entry += `
synchronous_standby_names = slave # standby servers that provide sync rep
                                  # comma-separated list of application_name
                                  # from standby(s); '*' = all`
	}

	//
	err = ioutil.WriteFile(file, []byte(entry), 0644)
	if err != nil {
		log.Error("[pg_config.configurePGConf] Failed to write to '%s'!\n%s\n", file, err)
	}

	return nil
}

//
func createRecovery() error {

	file := conf.DataDir+"recovery.conf"

	self := myself()
  other, err := Whois(otherRole(self))
  if err != nil {
  	log.Fatal("[pg_config.createRecovery] Unable to find another... Exiting!\n%s\n", err)
  	os.Exit(1)
  }

	//
	f, err := os.Create(file)
	if err != nil {
		log.Error("[pg_config.createRecovery] Failed to create '%s'!\n%s\n", file, err)
		return err
	}

	//
	entry := fmt.Sprintf(`
# -------------------------------------------------------
# PAGODA BOX
# -------------------------------------------------------

# IMPORTANT: this config file is dynamically generated by Pagoda Box for redundancy
# any changes made here will be overriden.

# When standby_mode is enabled, the PostgreSQL server will work as a standby. It
# tries to connect to the primary according to the connection settings
# primary_conninfo, and receives XLOG records continuously.
standby_mode = on
primary_conninfo = 'host=%s port=%d application_name=slave'

# restore_command specifies the shell command that is executed to copy log files
# back from archival storage. This parameter is *required* for an archive
# recovery, but optional for streaming replication. The given command satisfies
# the requirement without doing anything.
restore_command = 'exit 0'`, other.Ip, other.PGPort)

	//
	if _, err := f.WriteString(entry); err != nil {
		log.Error("[pg_config.createRecovery] Failed to write to '%s'!\n%s\n", file, err)
		return err
	}

	return nil
}

//
func destroyRecovery() {

	file := conf.DataDir+"recovery.conf"

	//
	err := os.Remove(file)
	if err != nil {
		log.Warn("[pg_config.destroyRecovery] No recovery.conf found at '%s'", file)
	}
}
