// pgconfig.go provides methods for configuring a node corresponding to if it is
// running a 'master' or 'slave' instance of postgres.

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

// configureHBAConf attempts to open the 'pg_hba.conf' file. Once open it will scan
// the file line by line looking for replication settings, and overwrite only those
// settings with the settings required for redundancy on Pagoda Box
func configureHBAConf() error {

	// get the role of the other node in the cluster that is running an instance
	// of postgresql
	self := Whoami()
	other, err := Whoisnot(self.CRole)
	if err != nil {
		log.Warn("[pg_config.configureHBAConf] Unable to find another!\n%s\n", err)
	}

	// open the pg_hba.conf
	file := conf.DataDir + "pg_hba.conf"
	f, err := os.Open(file)
	if err != nil {
		log.Error("[pg_config.configureHBAConf] Failed to open '%s'!\n%s\n", file, err)
		return err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	reFindConfigOption := regexp.MustCompile(`^\s*#?\s*(local|host)\s*(replication)`)
	entry := ""

	// scan the file line by line to build an 'entry' to be re-written back to the
	// file, skipping ('removing') any line that deals with redundancy.
	for scanner.Scan() {

		// dont care about submatches, just if the string matches
		if reFindConfigOption.FindString(scanner.Text()) == "" {
			entry += fmt.Sprintf("%s\n", scanner.Text())
		}
	}

	// if the other node is present, write redundancy into the 'entry' (otherwise
	// just leave it out)
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

	// write the 'entry' to the file
	err = ioutil.WriteFile(file, []byte(entry), 0644)
	if err != nil {
		log.Error("[pg_config.configureHBAConf] Failed to write to '%s'!\n%s\n", file, err)
		return err
	}

	return nil
}

// configurePGConf attempts to open the 'postgresql.conf' file. Once open it will
// scan the file line by line looking for replication settings, and overwrite only
// those settings with the settings required for redundancy on Pagoda Box
func configurePGConf(master bool) error {

	// open the postgresql.conf
	file := conf.DataDir + "postgresql.conf"
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	defer f.Close()

	reFindConfigOption := regexp.MustCompile(`^\s*#?\s*(listen_addresses|port|wal_level|archive_mode|archive_command|max_wal_senders|wal_keep_segments|hot_standby)\s*=\s*`)
	scanner := bufio.NewScanner(f)
	entry := ""

	// scan the file line by line to build an 'entry' to be re-written back to the
	// file, skipping ('removing') any lines that need to be manually configured
	for scanner.Scan() {

		submatch := reFindConfigOption.FindStringSubmatch(scanner.Text())

		//
		if submatch == nil {
			entry += fmt.Sprintf("%s\n", scanner.Text())
		}
	}

	// write manual configurations into an 'entry'
	entry += fmt.Sprintf(`
#------------------------------------------------------------------------------
# PAGODA BOX
#------------------------------------------------------------------------------

# these configuration options have been removed from their standard location and
# placed here so that Pagoda Box could override them with the neccessary values
# to configure redundancy.

# IMPORTANT: these settings will always be overriden when the server boots. They
# are set dynamically and so should never change.

listen_addresses = '0.0.0.0'      # what IP address(es) to listen on;
                                  # comma-separated list of addresses;
                                  # defaults to 'localhost'; use '*' for all
                                  # (change requires restart)
port = %d                     # (change requires restart)
wal_level = hot_standy            # minimal, archive, or hot_standby
                                  # (change requires restart)
archive_mode = on                 # allows archiving to be done
                                  # (change requires restart)
archive_command = 'exit 0'        # command to use to archive a logfile segment
                                  # placeholders: \%p = path of file to archive
                                  #               \%f = file name only
                                  # e.g. 'test ! -f /mnt/server/archivedir/\%f && cp \%p /mnt/server/archivedir/\%f'
max_wal_senders = 10              # max number of walsender processes
                                  # (change requires restart)
wal_keep_segments = 5000          # in logfile segments, 16MB each; 0 disables
hot_standby = on                  # "on" allows queries during recovery
                                  # (change requires restart)`, conf.PGPort)

	// if this node is currenty 'master' then write one additional configuration
	// into the 'entry'
	if master {
		entry += `
synchronous_standby_names = slave # standby servers that provide sync rep
                                  # comma-separated list of application_name
                                  # from standby(s); '*' = all`
	}

	// write 'entry' to the file
	err = ioutil.WriteFile(file, []byte(entry), 0644)
	if err != nil {
		log.Error("[pg_config.configurePGConf] Failed to write to '%s'!\n%s\n", file, err)
	}

	return nil
}

// createRecovery creates a 'recovery.conf' file with the necessary settings
// required for redundancy on Pagoda Box. This method is called on the node that
// is being configured to run the 'slave' instance of postgres
func createRecovery() error {

	file := conf.DataDir + "recovery.conf"

	self := Whoami()
	other, err := Whois(self.CRole)
	if err != nil {
		log.Fatal("[pg_config.createRecovery] Unable to find another... Exiting!\n%s\n", err)
		os.Exit(1)
	}

	// open/truncate the recover.conf
	f, err := os.Create(file)
	if err != nil {
		log.Error("[pg_config.createRecovery] Failed to create '%s'!\n%s\n", file, err)
		return err
	}

	// write manual configuration an 'entry'
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

	// write 'entry' to the file
	if _, err := f.WriteString(entry); err != nil {
		log.Error("[pg_config.createRecovery] Failed to write to '%s'!\n%s\n", file, err)
		return err
	}

	return nil
}

// destroyRecovery attempts to destroy the 'recovery.conf'. This method is called
// on a node that is being configured to run the 'master' instance of postgres
func destroyRecovery() {

	file := conf.DataDir + "recovery.conf"

	// remove 'recovery.conf'
	err := os.Remove(file)
	if err != nil {
		log.Warn("[pg_config.destroyRecovery] No recovery.conf found at '%s'", file)
	}
}
