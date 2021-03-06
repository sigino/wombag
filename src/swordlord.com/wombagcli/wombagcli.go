package main
/*-----------------------------------------------------------------------------
 **
 ** - Wombag -
 **
 ** the alternative, native backend for your Wallabag apps
 **
 ** Copyright 2017-18 by SwordLord - the coding crew - http://www.swordlord.com
 ** and contributing authors
 **
 ** This program is free software; you can redistribute it and/or modify it
 ** under the terms of the GNU Affero General Public License as published by the
 ** Free Software Foundation, either version 3 of the License, or (at your option)
 ** any later version.
 **
 ** This program is distributed in the hope that it will be useful, but WITHOUT
 ** ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 ** FITNESS FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License
 ** for more details.
 **
 ** You should have received a copy of the GNU Affero General Public License
 ** along with this program. If not, see <http://www.gnu.org/licenses/>.
 **
 **-----------------------------------------------------------------------------
 **
 ** Original Authors:
 ** LordEidi@swordlord.com
 ** LordLightningBolt@swordlord.com
 **
-----------------------------------------------------------------------------*/
import (
	"fmt"
	"os"
	"swordlord.com/wombag"
	"swordlord.com/wombagcli/cmd"
)

func main() {

	// Initialise env and params
	wombag.InitConfig()

	// Initialise database
	// todo: make sure database is working as expected, chicken out otherwise
	wombag.InitDatabase()
	defer wombag.CloseDB()

	// initialise the command structure
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)

		// yes, we deferred closing of the db, but that only works when ending with dignity
		wombag.CloseDB()
		os.Exit(1)
	}
}