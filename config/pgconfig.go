// Copyright (c) 2015 Nanobox Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

// pgconfig.go provides methods for configuring a node corresponding to if it is
// running a 'primary' or 'backup' instance of postgres.

package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	replicationRegex = regexp.MustCompile(`^\s*#?\s*(local|host)\s*(replication)`)
	overwriteRegex   = regexp.MustCompile(`^\s*#?\s*(listen_addresses|port|wal_level|archive_mode|archive_command|max_wal_senders|wal_keep_segments|hot_standby|synchronous_standby_names)\s*=\s*`)
)

// configureHBAConf attempts to open the 'pg_hba.conf' file. Once open it will scan
// the file line by line looking for replication settings, and overwrite only those
// settings with the settings required for redundancy on Yoke
func ConfigureHBAConf(ip string) error {

	// open the pg_hba.conf
	file := Conf.DataDir + "pg_hba.conf"
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	// scan the file line by line to build an 'entry' to be re-written back to the
	// file, skipping ('removing') any line that deals with redundancy.
	for scanner.Scan() {

		line := scanner.Text()

		// stop scanning if a special prefix is encountered.
		if strings.HasPrefix(line, "#~") {
			break
		}

		// dont care about submatches, just if the string matches, 'skipping' any lines
		// that are custom configurations
		if replicationRegex.MatchString(line) {
			_, err := fmt.Fprintf(f, "%s\n", line)
			if err != nil {
				return err
			}
		}
	}

	// add a replication connection into the hba.conf file so that data can be replicated
	// to other nodes
	_, err = fmt.Fprintf(f, `#~-----------------------------------------------------------------------------
# YOKE CONFIG
#------------------------------------------------------------------------------

# these configuration options have been removed from their standard location and
# placed here so that Yoke could override them with the neccessary values
# to configure redundancy.

# IMPORTANT: these settings will always be overriden when the server boots. They
# are set dynamically and so should never change.

host    replication     %s        %s/32            trust
`, Conf.SystemUser, ip)

	return err
}

// configurePGConf attempts to open the 'postgresql.conf' file. Once open it will
// scan the file line by line looking for replication settings, and overwrite only
// those settings with the settings required for redundancy
func ConfigurePGConf(ip string, port int) error {

	// open the postgresql.conf
	file := Conf.DataDir + "postgresql.conf"
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	// scan the file line by line to build an 'entry' to be re-written back to the
	// file, skipping ('removing') any lines that need to be manually configured
	for scanner.Scan() {

		line := scanner.Text()

		// stop scanning if a special prefix is encountered. This ensures there are
		// no duplicate Nanobox comment blocks
		if strings.HasPrefix(line, "#~") {
			break
		}

		// build the 'entry' from all lines that don't match the custom configurations
		if !overwriteRegex.MatchString(line) {
			_, err := fmt.Fprintf(f, "%s\n", line)
			if err != nil {
				return err
			}
		}
	}

	// write manual configurations into an 'entry'
	_, err = fmt.Fprintf(f, `#~-----------------------------------------------------------------------------
# YOKE CONFIG
#------------------------------------------------------------------------------

# these configuration options have been removed from their standard location and
# placed here so that Nanobox could override them with the neccessary values
# to configure redundancy.

# IMPORTANT: these settings will always be overriden when the server boots. They
# are set dynamically and so should never change.

listen_addresses = '%s'           # what IP address(es) to listen on;
                                  # comma-separated list of addresses;
                                  # defaults to 'localhost'; use '*' for all
                                  # (change requires restart)
port = %d                         # (change requires restart)
wal_level = hot_standby           # minimal, archive, or hot_standby
                                  # (change requires restart)
archive_mode = on                 # allows archiving to be done
                                  # (change requires restart)
archive_command = 'exit 0'        # command to use to archive a logfile segment
                                  # placeholders: \%p = path of file to archive
                                  #               \%f = file name only
                                  # e.g. 'test ! -f /mnt/server/archivedir/\%f && cp \%p /mnt/server/archivedir/\%f'
max_wal_senders = 10              # max number of walsender processes
                                  # (change requires restart)
wal_keep_segments = 16          	# in logfile segments, 16MB each; 0 disables
hot_standby = on                  # "on" allows queries during recovery
                                  # (change requires restart)
synchronous_standby_names = '*'   # standby servers that provide sync rep
                                  # comma-separated list of application_name
                                  # from standby(s); '*' = any
`, ip, port)

	return err
}

// createRecovery creates a 'recovery.conf' file with the necessary settings
// required for redundancy on Yoke. This method is called on the node that
// is being configured to run the 'backup' instance of postgres
func createRecovery(ip string, port int) error {

	file := Conf.DataDir + "recovery.conf"

	// open/truncate the recover.conf
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// write manual configuration an 'entry'
	_, err = fmt.Fprintf(f, `#~-----------------------------------------------------------------------------
# YOKE CONFIG
#------------------------------------------------------------------------------

# IMPORTANT: this config file is dynamically generated by Yoke for redundancy
# any changes made here will be overriden.

# When standby_mode is enabled, the PostgreSQL server will work as a standby. It
# tries to connect to the primary according to the connection settings
# primary_conninfo, and receives XLOG records continuously.
standby_mode = on
primary_conninfo = 'host=%s port=%d application_name=backup'

# restore_command specifies the shell command that is executed to copy log files
# back from archival storage. This parameter is *required* for an archive
# recovery, but optional for streaming replication. The given command satisfies
# the requirement without doing anything.
restore_command = 'exit 0'

# the presence of this file will stop this this node from recovering from the
# remote node.
trigger_file = '/data/var/db/postgresql/i-am-primary'
`, ip, port)

	return err
}
