package tablemodule
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
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"swordlord.com/wombag"
	"swordlord.com/wombag/model"
	"swordlord.com/wombagd/readability"
)

func NewFilter() Filter {
	filter := Filter{}
	filter.EntryId = 0
	filter.PerPage = 30
	filter.Page = 1
	filter.Starred = -1
	filter.Archive = -1
	return filter
}

type Filter struct {

	EntryId 	uint 	`form:"entry" json:"entry"` 	// Entry Id to filter for
	Tags 		string 	`form:"tags" json:"tags"` 		// tag1,tag2,tag3 	a comma-separated list of tags.
	Starred 	int 	`form:"starred" json:"starred"` // 1 or 0 	entry already starred
	Archive 	int 	`form:"archive" json:"archive"` // 1 or 0 	entry already archived

	Sort 		string 	`form:"sort" json:"sort"` 		// created or updated, default created  sort entries by date.
	Order 		string 	`form:"order" json:"order"` 		// asc, desc, default desc 	order of sort.
	Page 		int 	`form:"page" json:"page"` // 1, what page you want
	PerPage 	int 	`form:"perpage" json:"perpage"` // 30, results per page
	Since 		int 	`form:"since" json:"since"` // default 0, The timestamp since when you want entries updated.
}

func GetEntryTyped(entryId int) model.Entry {

	var entry model.Entry

	query := wombag.GetDB().First(&entry, entryId)

	tags := GetTagsPerEntry(uint(entryId))

	entry.Tags = tags

	if query.Error != nil {

	}

	return entry
}

func GetEntriesTyped(filter *Filter) []model.Entry {

	var rows []model.Entry

	// pagesize and start
	query := wombag.GetDB().Limit(filter.PerPage)
	query = query.Offset(filter.PerPage * (filter.Page - 1))

	if filter.Order == "desc" {
		query = query.Order("crt_dat DESC")
	} else {
		query = query.Order("crt_dat")
	}

	if filter.Starred >= 0 {
		query = query.Where("starred = ?", filter.Starred)
	}

	if filter.Archive >= 0 {
		query = query.Where("archive = ?", filter.Archive)
	}

	if filter.EntryId != 0 {
		query = query.Where("entry_id = ?", filter.EntryId)
	}

	if filter.Since >= 0 {

	}

	//expr := query.QueryExpr()
	//log.Printf("Query: %v\n", expr)
	query.Find(&rows)

	// todo: this is an ugly hack!
	for index, entry := range rows {

		tags := GetTagsPerEntry(uint(entry.EntryId))

		rows[index].Tags = tags
	}

	return rows
}

func GetEntries(withDetails bool) [][]string {

	filter := NewFilter()

	rows := GetEntriesTyped(&filter)

	var entries [][]string

	for _, entry := range rows {

		if withDetails {

			entries = append(entries, []string{ entry.Title, entry.Content, entry.CrtDat.Format("2006-01-02 15:04:05"), entry.UpdDat.Format("2006-01-02 15:04:05")})
		} else {

			entries = append(entries, []string{ entry.Title, entry.CrtDat.Format("2006-01-02 15:04:05"), entry.UpdDat.Format("2006-01-02 15:04:05")})
		}
	}

	return entries
}

func ListEntries() {

	entries := GetEntries(false)

	wombag.WriteTable([]string{"Title", "CrtDat", "UpdDat"}, entries)
}

func ExportEntries(file *os.File, ttl int) {

	log.Fatal("ExportEntries not implemented yet")
	/*
	entries := getEntries(false)

	for _, entry := range entries {

			//file.WriteString(entry[0] + "," + entry[1] + "," + entry + "\n")
	}
	*/
}

func AddEntry(Url string) (model.Entry, error) {

	response, err := http.Get(Url)
    if err != nil {
		log.Fatal(err)
		return model.Entry{}, err
	}
	defer response.Body.Close()

	content := ""
	title := ""
	//domain := response.Header.Get()

	if response.StatusCode == http.StatusOK {

		bodyBytes, err2 := ioutil.ReadAll(response.Body)
		if err2 != nil {
			log.Fatal(err2)
		}
		bodyString := string(bodyBytes)

		doc, err := readability.NewDocument(bodyString)
		if err != nil {
			log.Fatal(err)
		}

		content = doc.Content()
		title = doc.Title
	}

	db := wombag.GetDB()

	entry := model.Entry{URL: Url, Content: content, Title: title}
	retDB := db.Create(&entry)

	if retDB.Error != nil {
		log.Printf("Error with Entry %q: %s\n", Url, retDB.Error )
		log.Fatal(retDB.Error)
		return model.Entry{}, retDB.Error
	}

	fmt.Printf("Entry %s added.\n", Url)

	return entry, nil
}

func UpdateEntry(Id string, Starred bool, Archived bool, Title string) {

	db := wombag.GetDB()

	fields := make(map[string]interface{})

	fields["starred"] = Starred
	fields["archived"] = Archived

	if len(Title) > 0 {

		fields["title"] = Title
	}

	retDB := db.Model(&model.Entry{}).Where("entry_id=?", Id).Updates(fields)

	if retDB.Error != nil {
		log.Printf("Error with Entry %q: %s\n", Id, retDB.Error )
		log.Fatal(retDB.Error)
		return
	}

	fmt.Printf("Entry %s updated with Params: %s.\n", Id, Starred)

}

func DeleteEntry(EntryId uint) {

	db := wombag.GetDB()

	d := &model.Entry{}
	d.EntryId = EntryId

	ret := db.Delete(&d)

	if ret.Error != nil {

		fmt.Printf("Entry %s deletion resulted in an error %s.\n", EntryId, ret.Error)

	} else {

		fmt.Printf("Entry %s deleted.\n", EntryId)
	}
}
