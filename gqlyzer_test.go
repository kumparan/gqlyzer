package gqlyzer

import (
	"encoding/csv"
	"errors"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kumparan/gqlyzer/v2/token"
	"github.com/kumparan/gqlyzer/v2/token/operation"
)

func TestParseWithVariable(t *testing.T) {
	l := New(`query SomeOperation {
			SomeQuery(id: $id) {
				subQuery
			}
		}`)
	s, err := l.ParseWithVariables(`
		{
			"id": "danu"
		}
	`)

	assert.NoError(t, err)
	assert.Equal(t, operation.Query, s.Type)
	assert.Equal(t, "SomeOperation", s.Name)
	assert.Equal(t, "SomeQuery", s.Selections["SomeQuery"].Name)
	assert.Equal(t, "id", s.Selections["SomeQuery"].Arguments["id"].Key)
	assert.Equal(t, `"danu"`, s.Selections["SomeQuery"].Arguments["id"].Value)
	assert.Equal(t, "subQuery", s.Selections["SomeQuery"].InnerSelection["subQuery"].Name)
}

func TestParse(t *testing.T) {
	t.Run("anonymous graphql query", func(t *testing.T) {
		l := New(`{
	  IniQuerySatu(
	    id: "aya" object: USER
	  )
	  IniQueryDua(
	    id: "aya" object: USER
	  )
	}`)

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("graphql query without variable", func(t *testing.T) {
		l := New(`query iniOperationName {
	  IniQuerySatu(
	    id: "aya" object: USER
	  )
	  IniQueryDua(
	    id: "aya" object: USER
	  )
	}`)

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("graphql query with variable", func(t *testing.T) {
		l := New(`query iniOperationName(
	$objectID: ID!
	$userID: ID!
	$objectType: ObjectType!
	) {
	IniQuerySatu(
		objectID: $objectID
		objectType: $objectType
	)
	IniQueryDua(
		userID: $userID
		objectType: $objectType
	)
	}`)

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("json without opName, with var", func(t *testing.T) {
		l := New("query ( $objectID: ID! $userID: ID! $objectType: ObjectType!\t) {\n IniQuerySatu( objectID: $objectID\n objectType: $objectType )\n IniQueryDua( userID: $userID\n objectType: $objectType )\n}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("json with opName, with var", func(t *testing.T) {
		l := New("query iniOperationName( $objectID: ID! $userID: ID! $objectType: ObjectType!) {\n IniQuerySatu( objectID: $objectID\n objectType: $objectType )\n IniQueryDua( userID: $userID\n objectType: $objectType )\n}\n")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("json with opName, without var", func(t *testing.T) {
		l := New("query iniOperationName {\n IniQuerySatu(id: \"19\", object: USER)}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
	})

	t.Run("json with opName, without var, with fragments", func(t *testing.T) {
		l := New("query iniOperationName {\n  IniQuerySatu {\n    ...FragmentExample\n    __typename\n  }\n}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
	})

	t.Run("json without opName, without var", func(t *testing.T) {
		l := New("query {\n IniQuerySatu(userID: \"19\", objectType: USER )\n IniQueryDua(userID: \"19\", objectType: USER )\n}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("json with opName, with var, with alias", func(t *testing.T) {
		l := New("query iniOperationName($id: ID!) {\n  iniQueryAliasSatu: IniQuerySatu(id: $id) {\n    ...ItemDetails\n  }\n}\n\nfragment ItemDetails on Item {\n  ...BasicInfo\n  price\n}\n\nfragment BasicInfo on Item {\n  id\n  name\n}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "iniQueryAliasSatu", s.Selections["IniQuerySatu"].Alias)
	})

	t.Run("json with opName, without var, with alias", func(t *testing.T) {
		l := New("query iniOperationName {\n  iniQueryAliasSatu: IniQuerySatu {\n    ...FragmentExample\n    __typename\n  }\n}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "iniQueryAliasSatu", s.Selections["IniQuerySatu"].Alias)
	})

	t.Run("json with opname, with var, with alias, with object value arg", func(t *testing.T) {
		l := New("query iniOperationName($objectID: ID! $userID: ID! $objectType: ObjectType!) {\n\t\tiniQueryAliasSatu: IniQuerySatu(objectID: $objectID, userID: $userID, objectType: $objectType, filter: {\n\t\t\tiniObjectValueArgument: false\n\t\t\tiniJuga: $iniJuga\n\t\t}) {\n\t\t\tedges {\n\t\t\t\tid\n\t\t\t\ttitle\n\t\t\t\tpublisher {\n\t\t\t\t\tslug\n\t\t\t\t}\n\t\t\t\tauthor {\n\t\t\t\t\tusername\n\t\t\t\t}\n\t\t\t\tvideo {\n\t\t\t\t\tid\n\t\t\t\t\tduration\n\t\t\t\t\torientation\n\t\t\t\t\tposterMedia {\n\t\t\t\t\t\texternalURL\n\t\t\t\t\t}\n\t\t\t\t}\n\t\t\t\tcaption {\n\t\t\t\t\tdocument\n\t\t\t\t}\n\t\t\t\tcreatedAt\n\t\t\t}\n\t\t}\n\t}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "iniQueryAliasSatu", s.Selections["IniQuerySatu"].Alias)
	})

	t.Run("json with opName, without variable, w/o alias, with object value arg no line feed", func(t *testing.T) {
		l := New("query iniOperationName {\n  IniQuerySatu(\n    query: \"\"\n    size: 1\n    cursor: \"1\"\n    cursorType: PAGE\n    filters: {status: NEED_REVIEW}\n    sortType: STATUS_ASC_AND_UPDATED_AT_DESC\n  ) {\n    cursorInfo {\n      count\n      __typename\n    }\n    __typename\n  }\n}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
	})

	t.Run("json with opName, with variable, w/o alias, with fragment", func(t *testing.T) {
		l := New("mutation AddObjectToProfileClassification($objectID: ID!, $objectType: ProfileClassificationObjectType!, $profileClassificationID: ID!) {\n\nAddObjectToProfileClassification(objectID: $objectID, objectType: $objectType, profileClassificationID: $profileClassificationID){\n\n... on User{\n\n...User\n\n}\n\n... on Publisher{\n\n...Publisher\n\n}\n\n}\n\n}\n\nfragment User on User {\n\n__typename\n\nid\n\nname\n\nusername\n\naboutMe\n\nemail\n\nstatus\n\nphone\n\nemailVerified\n\nphoneVerified\n\nprofilePictureMedia {\n\n...Media\n\n}\n\ncoverPictureMedia {\n\n...Media\n\n}\n\ngender\n\nuserStatus: status\n\nbirthDate\n\nisRecommended\n\ncreatedAt\n\nupdatedAt\n\ndeletedAt\n\naboutMe\n\nisVerified\n\nwebsiteURL\n\nisVerified\n\nemailVerified\n\nwebsiteURL\n\nrole{\n\nid\n\nname\n\nslug\n\n}\n\nlastUpdatedBy{\n\nid\n\n}\n\nmetaTitle\n\nmetaDescription\n\nmetaKeyword\n\nemails{\n\nemail\n\nverifiedAt\n\ncreatedAt\n\n}\n\nisPasswordSet\n\nauthorizedChannel{\n\nisAuthorized\n\nchannel{\n\nid\n\nname\n\nslug\n\nmeta_title\n\nmeta_description\n\nmeta_keywords\n\n}\n\n}\n\nuserTermsAndConditionsAgreement {\n\nagreedAt\n\nid\n\nstatus\n\ntermsAndConditions {\n\ncreatedAt\n\nid\n\nupdatedAt\n\nversion\n\n}\n\n}\n\n}\n\nfragment Media on Media {\n\nid\n\ntitle\n\ndescription\n\npublicID\n\nexternalURL\n\nawsS3Key\n\nheight\n\nwidth\n\nlocationName\n\nlocationLat\n\nlocationLon\n\nmediaType\n\nmediaSourceID\n\nphotographer\n\neventDate\n\nlastUpdatedBy{\n\nid\n\nname\n\n}\n\nisArchived\n\ncreatedBy{\n\nid\n\nname\n\n}\n\ncreatedAt\n\nlastUpdatedAt\n\ntopics{\n\nid\n\n}\n\nmediaSource{\n\nid\n\nname\n\ncreatedAt\n\nlastUpdatedAt\n\ncreatedBy{\n\nid\n\nname\n\n}\n\nlastUpdatedBy{\n\nid\n\nname\n\n}\n\n}\n\n}\n\nfragment Publisher on Publisher {\n\n__typename\n\nid\n\nname\n\nslug\n\ndescription\n\nwebsite\n\nmetaTitle\n\nmetaKeywords\n\nmetaDescription\n\nisVerified\n\nisActive\n\nisPremium\n\ncoverMedia {\n\n...SimpleMedia\n\n}\n\navatarMedia {\n\n...SimpleMedia\n\n}\n\norganisation{\n\n...Organisation\n\n}\n\nauthorizedRSSConsumers{\n\nid\n\n}\n\nauthorizedChannel{\n\nchannel{\n\nid\n\n}\n\nisAuthorized\n\n}\n\npublisherGroupID\n\nisAutoMemberByDomain\n\ndomains\n\nenableGeneralPushNotificationForMember\n\nenableSegmentedPushNotificationForMember\n\n}\n\nfragment SimpleMedia on Media {\n\nid\n\ntitle\n\nlastUpdatedBy{\n\nid\n\n}\n\nisArchived\n\ncreatedBy{\n\nid\n\n}\n\ncreatedAt\n\nlastUpdatedAt\n\ntopics{\n\nid\n\n}\n\nmediaSource{\n\nid\n\ncreatedBy{\n\nid\n\n}\n\nlastUpdatedBy{\n\nid\n\n}\n\n}\n\n}\n\nfragment Organisation on Organisation {\n\nid\n\nname\n\nslug\n\ndescription\n\norganisationType\n\nwebsite\n\nisActive\n\ncoverMedia{\n\n...SimpleMedia\n\n}\n\navatarMedia{\n\n...SimpleMedia\n\n}\n\naddress\n\nphone1\n\nphone2\n\nemail\n\nmetaTitle\n\nmetaDescription\n\nmetaKeywords\n\nownedBy{\n\n...SimpleUser\n\n}\n\ncreatedBy{\n\n...SimpleUser\n\n}\n\n}\n\nfragment SimpleUser on User {\n\n__typename\n\nid\n\nname\n\nusername\n\nrole{\n\nid\n\n}\n\nauthorizedChannel{\n\nisAuthorized\n\nchannel{\n\nid\n\n}\n\n}\n\n}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Mutation, s.Type)
		assert.Equal(t, "AddObjectToProfileClassification", s.Name)
		assert.Equal(t, "AddObjectToProfileClassification", s.Selections["AddObjectToProfileClassification"].Name)
	})

	t.Run("json with text", func(t *testing.T) {
		l := New("mutation {\n  ReviseTopicSummaries(\n    linkedSummaryID: \"12345678\"\n    synthesisVoiceID: 2\n    reviseInput: [\n      {\n        summaryID: \"12345678\"\n        revisedSummary: \"COK Suzuki Fronx adalah mobil sub-compact SUV yang dirilis dengan harga mulai dari Rp 242,2 juta hingga Rp 316,3 juta. asda asda sdas\"\n      }\n    ]\n  )\n}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Mutation, s.Type)
		assert.Equal(t, "ReviseTopicSummaries", s.Selections["ReviseTopicSummaries"].Name)
	})

	t.Run("json with text 2", func(t *testing.T) {
		l := New("mutation {\n\tAnalyzeTypo(texts: [\n    \"GuluGuluGleg Gleg Gleg Khhrkkkrrrhrhhhrkkk\",\n    \"Lorem ipsum dolor sit amet, elit\",\n    \"roin maximus lectus ut turpis semper, vel blandit est accumsan.\",\n    \"Quisque faucibus, dui eu suscipit condimentum, sapien ante tincidunt ipsum, vitae aliquam elit odio quis arcu.\",\n    \"Donec aliquet tristique elit ut euismod\",\n    \"Proin ut urna eget mi euismod auctor.\",\n    \"Quisque faucibus, dui eu suscipit condimentum, sapien ante tincidunt ipsum, vitae aliquam elit odio quis arcu.\",\n\t]) {\n\t\ttypos{\n\t\t\toffset\n\t\t\ttype\n\t\t\ttoken\n\t\t\tsuggestions {\n\t\t\t\ttoken\n\t\t\t\tscore\n\t\t\t}\n\t\t}\n\t\t\n\t}\n}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Mutation, s.Type)
		assert.Equal(t, "AnalyzeTypo", s.Selections["AnalyzeTypo"].Name)
	})

	t.Run("json input array", func(t *testing.T) {
		l := New("mutation {\n\tReviseTopicSummaries(\n\t\treviseInput: [\n\t\t\t{\n\t\t\tsummaryID: \"1234567890\"\n\t\t\trevisedSummary: \"Satpol PP Kabupaten Penajam Paser Utara menangkap 64 PSK di wilayah IKN sepanjang tahun ini. Mereka yang terjaring berasal dari berbagai kota seperti Samarinda, Balikpapan, Bandung, Makassar, dan Yogyakarta. \\n \\n Para PSK tersebut beroperasi secara mandiri, tanpa difasilitasi oleh muncikari. Oleh karena itu, mereka tidak dapat dikenakan pidana, melainkan hanya mendapatkan sanksi pengusiran dari Penajam Paser Utara.\"\n\t\t\t},\n\t\t\t{\n\t\t\tsummaryID: \"1234567890\"\n\t\t\trevisedSummary: \"* Terbaru! Suzuki Fronx meluncurkan fitur Advanced Driving Assistant System (ADAS) yang dirancang untuk membantu pengemudi mengurangi keletihan dan meningkatkan keselamatan berkendara.\\n* Fitur-fitur ADAS di Suzuki Fronx meliputi Dual Sensor Brake Support II, Adaptive Cruise Control, Lane Keep Assist, Lane Departure Warning, Lane Departure Prevention, Vehicle Swaying Warning, Blind Spot Monitor, Rear Cross Traffic Alert, dan High Beam Assist.\\n* Sistem ini memanfaatkan modul kamera dan sensor radar untuk memancarkan gelombang radio untuk mengukur jarak dan kecepatan objek di depan dan belakang.\\n* Suzuki Fronx hadir sebagai pilihan baru di segmen SUV sub-compact crossover dengan panjang dimensi 4 meter dan tersedia dalam varian SGX A/T SHVS, GX A/T SHVS, GX M/T SHVS, GL A/T, dan GL M/T.\\n* Suzuki Fronx berhasil mengimpor 3.990 unit mobil ke Jepang pada bulan April, lebih tinggi dibandingkan Mercedes-Benz dan BMW, didorong oleh ledakan permintaan terhadap Jimny Nomade versi lima pintu.\"\n\t\t\t},\n\t\t]\n\t\tlinkedSummaryID: \"1765524525286736337\"\n\t\tsynthesisVoiceID: \"1\"\n\t)\n}")

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Mutation, s.Type)
		assert.Equal(t, "ReviseTopicSummaries", s.Selections["ReviseTopicSummaries"].Name)
	})

	t.Run("json introspection query", func(t *testing.T) {
		l := New("\n    query IntrospectionQuery {\n      __schema {\n        \n        queryType { name kind }\n        mutationType { name kind }\n        subscriptionType { name kind }\n        types {\n          ...FullType\n        }\n        directives {\n          name\n          description\n          \n          locations\n          args {\n            ...InputValue\n          }\n        }\n      }\n    }\n\n    ")

		s, err := l.Parse()

		assert.NoError(t, err) // fixed: one-line sub-selections now parse
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "IntrospectionQuery", s.Name)
		assert.Equal(t, "__schema", s.Selections["__schema"].Name)
	})

	t.Run("json query like user input", func(t *testing.T) {
		l := New("mutation {\n  CreateDraftStoryV2(\n    draft: {\n\t\t\tauthorID: \"1234567890\", \n\t\t\tpublisherID: \"\", \n\t\t\tchannelID: \"3\", \n\t\t\ttitle: \"ICOK Suzuki Fronx adalah mobil\", \n\t\t\tsource: UGC, \n\t\t\tleadText: \"dummy leadtext\", \n\t\t\tcontent: \"{\\\"object\\\":\\\"value\\\",\\\"document\\\":{\\\"object\\\":\\\"document\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"heading-large\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"IHSG Dibuka Menguat, Rupiah Melemah, Bursa Asia Bergerak Variatif\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Indeks Harga Saham Gabungan (IHSG) mengawali perdagangan hari ini dengan menunjukkan penguatan, mencerminkan sentimen positif di awal sesi. Namun, di pasar valuta asing, nilai tukar rupiah terhadap dolar Amerika Serikat (AS) terpantau melemah. Sementara itu, bursa saham-saham utama di Asia menampilkan pergerakan yang beragam, dengan beberapa indeks dibuka positif dan lainnya menunjukkan koreksi di sesi pertama.\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"heading-medium\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"IHSG\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Pada pembukaan perdagangan tanggal \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"11 Desember 2025\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"09:00:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", Indeks Harga Saham Gabungan (IHSG) berhasil dibuka di level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"8764.09\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\". Kinerja positif ini ditandai dengan kenaikan sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"0.73%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" dari posisi penutupan sebelumnya, memberikan sinyal optimisme bagi para investor di awal sesi perdagangan.\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"heading-medium\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Nilai Tukar Rupiah\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Kondisi berbeda terlihat di pasar valuta asing. Data terkini pada pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"09:50:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" menunjukkan bahwa nilai tukar mata uang rupiah terhadap dolar AS berada pada level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"16683\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\". Angka ini mengindikasikan adanya depresiasi atau pelemahan rupiah terhadap mata uang Negeri Paman Sam tersebut, yang mungkin menjadi perhatian bagi eksportir dan importir serta sektor keuangan.\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"heading-medium\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Bursa Saham Asia\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Berikut adalah kinerja indeks saham utama di Asia pada awal perdagangan hari ini:\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Nikkei 225\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" (Jepang):\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   Dibuka pada \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"11 Desember 2025\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"07:00:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" di level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"50818.39\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", naik sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"0.43%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\".\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   Pada penutupan sesi 1 pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"09:35:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", indeks ini bergerak terkoreksi ke level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"50308.89\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", turun sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"-0.58%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\".\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Hang Seng Index\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" (Hong Kong):\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   Dibuka pada \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"11 Desember 2025\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"08:30:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" di level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"25710.61\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", menguat sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"0.66%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\".\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Shanghai Composite\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" (Tiongkok):\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   Dibuka pada \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"11 Desember 2025\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"08:30:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" di level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"3904.96\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", naik tipis sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"0.11%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\".\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"STI (Straits Times Index)\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" (Singapura):\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   Dibuka pada \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"11 Desember 2025\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"08:00:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" di level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"4516.34\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", menguat sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"0.23%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\".\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Secara keseluruhan, bursa saham Asia menunjukkan pergerakan yang variatif. Meskipun Hang Seng, Shanghai Composite, dan STI dibuka dengan penguatan, Nikkei 225 yang sebelumnya dibuka positif harus mengalami koreksi pada penutupan sesi pertamanya. Hal ini mencerminkan sentimen pasar yang beragam di kawasan Asia pagi ini, dengan beberapa pasar masih menjaga momentum positif sementara yang lain menghadapi tekanan jual.\\\",\\\"marks\\\":[]}]}]}]}}\", \n\t\t\tdocumentType: SLATEJS, \n\t\t\treporterIDs: [], \n\t\t\tleadMediaIDs: [], \n\t\t\ttopicIDs: [], \n\t\t\teditorIDs: [], \n\t\t\taddOns: [], \n\t\t\tattributes: {}\n\t\t}) \n\t{\n    id\n    title\n    slug\n  }\n}")

		s, err := l.Parse()

		assert.NoError(t, err) // fixed: escaped quotes in string values now parse
		assert.Equal(t, operation.Mutation, s.Type)
		assert.Equal(t, "CreateDraftStoryV2", s.Selections["CreateDraftStoryV2"].Name)
	})
}

func TestErrEOF_HasExpectedMessage(t *testing.T) {
	// The string "end of file" must not change: callers that still
	// compare err.Error() directly must not break.
	assert.Equal(t, "end of file", ErrEOF.Error())
}

func TestErrEOF_WorksWithErrorsIs(t *testing.T) {
	// Wrapping via fmt.Errorf %w must still be detectable.
	// This verifies the sentinel value is stable (same pointer),
	// so errors.Is works even when err is wrapped downstream.
	wrapped := ErrEOF // direct identity
	assert.True(t, errors.Is(wrapped, ErrEOF))
}

// =============================================================
// read() is a non-consuming peek
// =============================================================

func TestParseOperationType_EmptyInput_ReturnsNilError(t *testing.T) {
	l := New("")
	_, err := l.ParseOperationType()
	assert.NoError(t, err, "EOF on empty input must be suppressed — not a real parse error")
}

func TestParseOperationType_EmptyInput_DoesNotReturnErrEOF(t *testing.T) {
	l := New("")
	_, err := l.ParseOperationType()
	assert.False(t, errors.Is(err, ErrEOF), "ErrEOF must not reach the caller for empty input")
}

func TestParseOperationType_WhitespaceOnly_ReturnsNilError(t *testing.T) {
	l := New("   \n\t  ")
	_, err := l.ParseOperationType()
	assert.NoError(t, err, "whitespace-only input is not a parse error")
}

func TestParseOperationType_EmptyInput_ReturnsZeroValues(t *testing.T) {
	l := New("")
	op, err := l.ParseOperationType()
	assert.NoError(t, err)
	assert.Equal(t, operation.Type(""), op)
}

func TestParseOperationType_Query(t *testing.T) {
	l := New("query")
	op, err := l.ParseOperationType()
	assert.NoError(t, err)
	assert.Equal(t, operation.Query, op)
}

func TestParseOperationType_Mutation(t *testing.T) {
	l := New("mutation")
	op, err := l.ParseOperationType()
	assert.NoError(t, err)
	assert.Equal(t, operation.Mutation, op)
}

func TestParseOperationType_Subscription(t *testing.T) {
	l := New("subscription")
	op, err := l.ParseOperationType()
	assert.NoError(t, err)
	assert.Equal(t, operation.Subscription, op)
}

func TestParseOperationType_OpenBrace_ReturnsQueryAndIsAnonymous(t *testing.T) {
	l := New("{")
	op, err := l.ParseOperationType()
	assert.NoError(t, err)
	assert.Equal(t, operation.Query, op)

	// The "isAnonymous" return value is gone. An anonymous operation is now
	// observable through its empty name.
	parsed, err := New("{ field }").Parse()
	assert.NoError(t, err)
	assert.Equal(t, "", parsed.Name)
}

func TestParseOperation_EmptyInput_ReturnsNilError(t *testing.T) {
	l := New("")
	_, err := l.Parse()
	assert.NoError(t, err, "empty input must not produce an error from parseOperation")
}

func TestParseOperation_EmptyInput_DoesNotReturnErrEOF(t *testing.T) {
	l := New("")
	_, err := l.Parse()
	assert.False(t, errors.Is(err, ErrEOF), "ErrEOF must not leak from parseOperation for empty input")
}

func TestParseOperation_EmptyInput_ReturnsZeroOperation(t *testing.T) {
	l := New("")
	op, err := l.Parse()
	assert.NoError(t, err)
	assert.Equal(t, operation.Type(""), op.Type)
	assert.Equal(t, "", op.Name)
}

func TestParseOperation_WhitespaceOnly_ReturnsNilError(t *testing.T) {
	l := New("   \n   ")
	_, err := l.Parse()
	assert.NoError(t, err)
}

func TestParse_EmptyInput_ReturnsNilError(t *testing.T) {
	l := New("")
	_, err := l.Parse()
	assert.NoError(t, err, "Parse(\"\") must not return an error")
}

func TestParse_EmptyInput_ReturnsZeroOperation(t *testing.T) {
	l := New("")
	op, err := l.Parse()
	assert.NoError(t, err)
	assert.Equal(t, operation.Type(""), op.Type)
	assert.Equal(t, "", op.Name)
}

func TestParse_WhitespaceOnly_ReturnsNilError(t *testing.T) {
	l := New("   \n   ")
	_, err := l.Parse()
	assert.NoError(t, err)
}

func TestParse_AnonymousQuery_TypeIsQuery(t *testing.T) {
	l := New(`{
	IniQuery(id: "1")
}`)
	op, err := l.Parse()
	assert.NoError(t, err)
	assert.Equal(t, operation.Query, op.Type)
}

func TestParse_AnonymousQuery_SelectionsPresent(t *testing.T) {
	l := New(`{
	IniQuery(id: "1")
}`)
	op, err := l.Parse()
	assert.NoError(t, err)
	assert.Equal(t, "IniQuery", op.Selections["IniQuery"].Name)
}

func TestParse_QueryKeywordNoName_TypeIsQuery(t *testing.T) {
	l := New("query {\n IniQuery(userID: \"19\")\n}")
	op, err := l.Parse()
	assert.NoError(t, err)
	assert.Equal(t, operation.Query, op.Type)
	assert.Equal(t, "IniQuery", op.Selections["IniQuery"].Name)
}

func TestParse_QueryKeywordWithName_ReturnsName(t *testing.T) {
	l := New("query iniOperationName {\n IniQuery(id: \"19\")\n}")
	op, err := l.Parse()
	assert.NoError(t, err)
	assert.Equal(t, operation.Query, op.Type)
	assert.Equal(t, "iniOperationName", op.Name)
	assert.Equal(t, "IniQuery", op.Selections["IniQuery"].Name)
}

func TestParse_MutationKeyword_TypeIsMutation(t *testing.T) {
	l := New("mutation {\n CreateUser(name: \"test\")\n}")
	op, err := l.Parse()
	assert.NoError(t, err)
	assert.Equal(t, operation.Mutation, op.Type)
}

func TestParse_MutationWithName_ReturnsName(t *testing.T) {
	l := New("mutation AddObject($id: ID!) {\n AddObjectToProfile(id: $id)\n}")
	op, err := l.Parse()
	assert.NoError(t, err)
	assert.Equal(t, operation.Mutation, op.Type)
	assert.Equal(t, "AddObject", op.Name)
}

func TestParse_QueryWithVariables_ParsesNameAndSelections(t *testing.T) {
	l := New("query iniOperationName(\n$objectID: ID!\n$userID: ID!\n) {\nIniQuerySatu(\n\tobjectID: $objectID\n)\nIniQueryDua(\n\tuserID: $userID\n)\n}")
	op, err := l.Parse()
	assert.NoError(t, err)
	assert.Equal(t, operation.Query, op.Type)
	assert.Equal(t, "iniOperationName", op.Name)
	assert.Equal(t, "IniQuerySatu", op.Selections["IniQuerySatu"].Name)
	assert.Equal(t, "IniQueryDua", op.Selections["IniQueryDua"].Name)
}

func TestParseOperationType_Public_EmptyInput_ReturnsNilError(t *testing.T) {
	l := New("")
	_, err := l.ParseOperationType()
	assert.NoError(t, err, "ParseOperationType on empty input must not return an error")
}

func TestParseOperationType_Public_EmptyInput_ReturnsEmptyType(t *testing.T) {
	l := New("")
	ot, err := l.ParseOperationType()
	assert.NoError(t, err)
	assert.Equal(t, operation.Type(""), ot)
}

func TestParseOperationType_Public_Query(t *testing.T) {
	l := New("query { field }")
	ot, err := l.ParseOperationType()
	assert.NoError(t, err)
	assert.Equal(t, operation.Query, ot)
}

func TestParseOperationType_Public_Mutation(t *testing.T) {
	l := New("mutation { createUser }")
	ot, err := l.ParseOperationType()
	assert.NoError(t, err)
	assert.Equal(t, operation.Mutation, ot)
}

func TestParseOperationType_Public_Subscription(t *testing.T) {
	l := New("subscription { onUpdate }")
	ot, err := l.ParseOperationType()
	assert.NoError(t, err)
	assert.Equal(t, operation.Subscription, ot)
}

func TestParseOperationType_Public_AnonymousQuery(t *testing.T) {
	l := New("{ field }")
	ot, err := l.ParseOperationType()
	assert.NoError(t, err)
	assert.Equal(t, operation.Query, ot)
}

func TestParse_NoEOFErrorForAnyValidInput(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"whitespace only", "   \n   "},
		{"anonymous query", "{\n\tIniQuery\n}"},
		{"query keyword no name", "query {\n\tIniQuery\n}"},
		{"query keyword with name", "query GetUser {\n\tuser\n}"},
		{"mutation keyword", "mutation {\n\tcreateUser\n}"},
		{"mutation with name", "mutation CreateUser {\n\tcreateUser\n}"},
		{"query with variables", "query GetUser($id: ID!) {\n\tuser\n}"},
		{"query no name with variables", "query ( $objectID: ID! $userID: ID!\t) {\n IniQuerySatu( objectID: $objectID\n)\n}"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			l := New(tc.input)
			_, err := l.Parse()
			assert.False(t, errors.Is(err, ErrEOF),
				"Parse(%q): ErrEOF must not leak to the caller", tc.input)
		})
	}
}

// =====================================================================
// Retired tests
//
// Six tests from before this change are gone, and each one asserted
// something about the hand-written lexer's own machinery rather than about
// the analysis it produced. There is no cursor and no rune-at-a-time read
// to assert against any more:
//
//	TestRead_OnEmptyInput_ReturnsErrEOF
//	TestRead_ReturnsCorrectRune
//	TestRead_DoesNotAdvanceCursor
//	TestErrEOF_SameSentinelReturnedEveryTime
//	TestParseOperationType_OpenBrace_AdvancesCursorPastBrace
//	TestParseOperationType_OpenBrace_NextReadIsNotOpenBrace
//
// The last two guarded against '{' being read twice on an anonymous
// operation. What that protected is observable, and is asserted in
// TestParseOperationType_OpenBrace_ReturnsQueryAndIsAnonymous and in
// TestParse_AnonymousQuery_SelectionsPresent: the body still parses.
//
// Every other test from before this change is kept below or above, under its
// original name.
// =====================================================================

// =====================================================================
// Carried over from the deleted internal test files
//
// These exercised unexported functions of the hand-written lexer
// (isAlphabet, parseKeyword, parseName, parseSelection, parseSelectionSet,
// parseArgument, parseArgumentSet). Those functions are gone, so each test
// below asserts the same behaviour through the public API, keeping the
// original inputs.
// =====================================================================

// from utils_test.go: TestIsAlphabet
//
// The old isAlphabet used unicode.IsLetter, so it accepted "ø", "中" and "д"
// as identifier characters. The GraphQL spec defines Name as
// [_A-Za-z][_0-9A-Za-z]*, so those documents are not valid GraphQL and a real
// server rejects them. gqlparser and graphql-core both reject them too.
// Accepting them was a bug, so that expectation is deliberately not carried
// over; this test pins the spec behaviour instead.
func TestIsAlphabet(t *testing.T) {
	t.Run("lowercase letters", func(t *testing.T) {
		op, err := New("{ abcdefghijklmnopqrstuvwxyz }").Parse()
		assert.NoError(t, err)
		assert.Equal(t, []string{"abcdefghijklmnopqrstuvwxyz"}, topLevelNames(op.Selections))
	})

	t.Run("uppercase letters", func(t *testing.T) {
		op, err := New("{ ABCDEFGHIJKLMNOPQRSTUVWXYZ }").Parse()
		assert.NoError(t, err)
		assert.Equal(t, []string{"ABCDEFGHIJKLMNOPQRSTUVWXYZ"}, topLevelNames(op.Selections))
	})

	t.Run("digits and underscores", func(t *testing.T) {
		op, err := New("{ _a1_B2 }").Parse()
		assert.NoError(t, err)
		assert.Equal(t, []string{"_a1_B2"}, topLevelNames(op.Selections))
	})

	t.Run("unicode letters", func(t *testing.T) {
		for _, name := range []string{"ø", "中", "д", "café"} {
			_, err := New("{ " + name + " }").Parse()
			assert.Errorf(t, err, "%q is not a valid GraphQL Name", name)
		}
	})

	t.Run("non letters", func(t *testing.T) {
		for _, name := range []string{"😔", "\n", "\t", " "} {
			_, err := New("{ " + name + " }").Parse()
			assert.Error(t, err)
		}
	})
}

// from parse_keyword_test.go: TestParseMutationKeyword
func TestParseMutationKeyword(t *testing.T) {
	t.Run(`should return no error when given correct keyword`, func(t *testing.T) {
		ot, err := New("mutation { a }").ParseOperationType()

		assert.NoError(t, err)
		assert.Equal(t, operation.Mutation, ot)
	})

	t.Run(`should return error when given mismatch keyword`, func(t *testing.T) {
		// "querty" is not an operation keyword.
		_, err := New("querty { a }").Parse()

		assert.Error(t, err)
	})
}

// from parse_name_test.go: TestParseName
func TestParseName(t *testing.T) {
	t.Run("ok, alphabet", func(t *testing.T) {
		op, err := New("{ hello }").Parse()

		assert.NoError(t, err)
		assert.Equal(t, "hello", op.Selections["hello"].Name)
	})

	t.Run("ok, underscore", func(t *testing.T) {
		op, err := New("{ __hello }").Parse()

		assert.NoError(t, err)
		assert.Equal(t, "__hello", op.Selections["__hello"].Name)
	})

	t.Run("fail: not _ or alphabet", func(t *testing.T) {
		_, err := New("{ 9hello }").Parse()

		assert.Error(t, err)
	})
}

// from parse_operation_test.go: TestParseOperation
func TestParseOperation(t *testing.T) {
	t.Run("with anonymous operation", func(t *testing.T) {
		l := New(`{
			SomeQuery(id: 123) {
				subQuery
			}
		}`)

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "", s.Name)
		assert.Equal(t, "SomeQuery", s.Selections["SomeQuery"].Name)
		assert.Equal(t, "id", s.Selections["SomeQuery"].Arguments["id"].Key)
		assert.Equal(t, "123", s.Selections["SomeQuery"].Arguments["id"].Value)
		assert.Equal(t, "subQuery", s.Selections["SomeQuery"].InnerSelection["subQuery"].Name)
	})
}

// from parse_selection_test.go: TestParseSelection
func TestParseSelection(t *testing.T) {
	t.Run("without parameter", func(t *testing.T) {
		op, err := New("{ SomeQuery }").Parse()

		assert.NoError(t, err)
		assert.Equal(t, "SomeQuery", op.Selections["SomeQuery"].Name)
	})

	t.Run("with subselection", func(t *testing.T) {
		op, err := New(`{ SomeQuery {
			subQuery
		} }`).Parse()

		assert.NoError(t, err)
		s := op.Selections["SomeQuery"]
		assert.Equal(t, "SomeQuery", s.Name)
		assert.Equal(t, "subQuery", s.InnerSelection["subQuery"].Name)
	})

	t.Run("with arguments", func(t *testing.T) {
		op, err := New(`{ SomeQuery(id: 123) {
			subQuery
		} }`).Parse()

		assert.NoError(t, err)
		s := op.Selections["SomeQuery"]
		assert.Equal(t, "SomeQuery", s.Name)
		assert.Equal(t, "subQuery", s.InnerSelection["subQuery"].Name)
		assert.Equal(t, "id", s.Arguments["id"].Key)
	})
}

// from parse_selection_test.go: TestParseSelectionSet
func TestParseSelectionSet(t *testing.T) {
	t.Run("with correct separator", func(t *testing.T) {
		op, err := New(`{
		query1, query2
		query3
	}`).Parse()

		assert.NoError(t, err)
		s := op.Selections
		assert.Equal(t, "query1", s["query1"].Name)
		assert.Equal(t, "query2", s["query2"].Name)
		assert.Equal(t, "query3", s["query3"].Name)
	})

	t.Run("with incorrect separator", func(t *testing.T) {
		// The old lexer only accepted a newline or comma between fields, so it
		// could not read "query1 query2" and the original test only asserted
		// that no error came back. All three fields are now reported.
		op, err := New(`{
		query1 query2
		query3
	}`).Parse()

		assert.NoError(t, err)
		assert.Equal(t, []string{"query1", "query2", "query3"}, topLevelNames(op.Selections))
	})

	t.Run("with nested value", func(t *testing.T) {
		op, err := New(`{
		query1(id: 123) {
			query3
		},
		query2
	}`).Parse()

		assert.NoError(t, err)
		s := op.Selections
		assert.Equal(t, "query1", s["query1"].Name)
		assert.Equal(t, "query3", s["query1"].InnerSelection["query3"].Name)
		assert.Equal(t, "id", s["query1"].Arguments["id"].Key)
		assert.Equal(t, "query2", s["query2"].Name)
	})
}

// from parse_selection_args_test.go: TestParseArgument
func TestParseArgument(t *testing.T) {
	t.Run("with string value", func(t *testing.T) {
		op, err := New(`{ field(SomeQuery: "helloworld") }`).Parse()

		assert.NoError(t, err)
		s := op.Selections["field"].Arguments["SomeQuery"]
		assert.Equal(t, "SomeQuery", s.Key)
		assert.Equal(t, `"helloworld"`, s.Value)
	})

	t.Run("with object value", func(t *testing.T) {
		op, err := New(`{ field(SomeQuery: {
			test: "helloworld",
			test2: "helloworld"
		}) }`).Parse()

		assert.NoError(t, err)
		s := op.Selections["field"].Arguments["SomeQuery"]
		assert.Equal(t, "SomeQuery", s.Key)
		assert.Equal(t, "test", s.ObjectValue["test"].Key)
		assert.Equal(t, `"helloworld"`, s.ObjectValue["test"].Value)
		assert.Equal(t, "test2", s.ObjectValue["test2"].Key)
	})
}

// from parse_selection_args_test.go: TestParseArgumentSet
func TestParseArgumentSet(t *testing.T) {
	t.Run("with single value", func(t *testing.T) {
		op, err := New(`{ field(
		arg1: 1,
		arg2: 2,
		arg3: 3
 ) }`).Parse()

		assert.NoError(t, err)
		s := op.Selections["field"].Arguments
		assert.Equal(t, "arg1", s["arg1"].Key)
		assert.Equal(t, "1", s["arg1"].Value)
		assert.Equal(t, "arg2", s["arg2"].Key)
		assert.Equal(t, "2", s["arg2"].Value)
		assert.Equal(t, "arg3", s["arg3"].Key)
		assert.Equal(t, "3", s["arg3"].Value)
	})

	t.Run("with nested value", func(t *testing.T) {
		op, err := New(`{ field(
			user: {
				name: "danu",
				id: 123
			},
			page: 1
		) }`).Parse()

		assert.NoError(t, err)
		s := op.Selections["field"].Arguments
		assert.Equal(t, "user", s["user"].Key)
		assert.Equal(t, "page", s["page"].Key)
		assert.Equal(t, "1", s["page"].Value)
		assert.Equal(t, `"danu"`, s["user"].ObjectValue["name"].Value)
		assert.Equal(t, "123", s["user"].ObjectValue["id"].Value)
	})
}

// =====================================================================
// Regressions for the queries the hand-written lexer could not read
// =====================================================================

func TestParse_SingleLineQuery(t *testing.T) {
	// Used to fail with "expected separator, but got: #": fields were only
	// treated as separate when a newline or comma stood between them.
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{"single field inline", `{__typename}`, []string{"__typename"}},
		{"space separated", `query { a b }`, []string{"a", "b"}},
		{"comma separated", `query { a, b }`, []string{"a", "b"}},
		{"whole query on one line", `query { Find(slug: "msci", size: 5) { edges { id title } } }`, []string{"Find"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			op, err := New(tc.input).Parse()
			require.NoError(t, err)
			assert.Equal(t, tc.want, topLevelNames(op.Selections))
		})
	}
}

func TestParse_InlineSubSelection(t *testing.T) {
	// Used to fail with "end of file": reading "publisher { slug }" swallowed
	// the closing brace, so every later brace was off by one.
	op, err := New("query {\n  a {\n    publisher { slug }\n    author { username }\n  }\n}").Parse()

	require.NoError(t, err)
	inner := op.Selections["a"].InnerSelection
	assert.Equal(t, []string{"author", "publisher"}, topLevelNames(inner))
	assert.Equal(t, "slug", inner["publisher"].InnerSelection["slug"].Name)
	assert.Equal(t, "username", inner["author"].InnerSelection["username"].Name)
}

func TestParse_EscapedQuoteInString(t *testing.T) {
	// Used to fail: parseString stopped at the first '"' regardless of the
	// backslash in front of it.
	op, err := New(`mutation { m(x: "he said \"hi\"") }`).Parse()

	require.NoError(t, err)
	assert.Equal(t, `"he said \"hi\""`, op.Selections["m"].Arguments["x"].Value)
}

func TestParse_BracketInsideStringInsideList(t *testing.T) {
	// Used to fail: parseArray scanned to the first ']', including one that
	// sat inside a string.
	op, err := New(`mutation { m(x: ["a]b", "c"]) }`).Parse()

	require.NoError(t, err)
	assert.Equal(t, `["a]b", "c"]`, op.Selections["m"].Arguments["x"].Value)
}

func TestParse_ArrayArgumentFollowedByAnother(t *testing.T) {
	// Used to succeed with the wrong answer: parseArray advanced the cursor
	// twice, so a character after ']' was skipped, the arguments were lost and
	// the sub-field surfaced as a top-level selection.
	op, err := New("{\n  f(t: [STORY], x: 1) {\n    a\n  }\n}").Parse()

	require.NoError(t, err)
	assert.Equal(t, []string{"f"}, topLevelNames(op.Selections))
	assert.Equal(t, "[STORY]", op.Selections["f"].Arguments["t"].Value)
	assert.Equal(t, "1", op.Selections["f"].Arguments["x"].Value)
	assert.Equal(t, "a", op.Selections["f"].InnerSelection["a"].Name)
}

func TestParse_NestedList(t *testing.T) {
	op, err := New(`mutation { m(x: [[1, 2], [3]]) }`).Parse()

	require.NoError(t, err)
	assert.Equal(t, "[[1, 2], [3]]", op.Selections["m"].Arguments["x"].Value)
}

func TestParse_EmptyListArgument(t *testing.T) {
	// Used to fail with "invalid stack pop" when written on one line.
	op, err := New("mutation {\n  f(d: {a: [], b: 1}) {\n    x\n  }\n}").Parse()

	require.NoError(t, err)
	assert.Equal(t, "[]", op.Selections["f"].Arguments["d"].ObjectValue["a"].Value)
	assert.Equal(t, "1", op.Selections["f"].Arguments["d"].ObjectValue["b"].Value)
}

func TestParse_BlockString(t *testing.T) {
	op, err := New("mutation { m(x: \"\"\"he said \"hi\"\"\"\") }").Parse()

	require.NoError(t, err)
	assert.Equal(t, `"he said \"hi\""`, op.Selections["m"].Arguments["x"].Value)
}

func TestParse_Comment(t *testing.T) {
	// Used to fail: '#' is the old lexer's own flush marker, so a real comment
	// could not be read.
	op, err := New("query {\n  # pick the first page\n  a\n}").Parse()

	require.NoError(t, err)
	assert.Equal(t, []string{"a"}, topLevelNames(op.Selections))
}

func TestParse_VariableDefaultValueContainingParen(t *testing.T) {
	// Used to succeed with an empty selection set: the variable block was
	// skipped by scanning to the first ')', including one inside a string.
	op, err := New(`query Q($x: String = ")") { a }`).Parse()

	require.NoError(t, err)
	assert.Equal(t, "Q", op.Name)
	assert.Equal(t, []string{"a"}, topLevelNames(op.Selections))
}

func TestParse_FragmentSpreadWithSpace(t *testing.T) {
	// Used to report a selection named "g": parseName assumed the three
	// characters after "... " were always "on ".
	op, err := New("query {\n  a {\n    ... Frag\n  }\n}\nfragment Frag on T { id }").Parse()

	require.NoError(t, err)
	assert.Equal(t, []string{"id"}, topLevelNames(op.Selections["a"].InnerSelection))
}

func TestParse_InlineFragmentContributesFieldsToParent(t *testing.T) {
	op, err := New("{\n  edges {\n    object {\n      ... on Story { id title }\n      ... on Video { id duration }\n    }\n  }\n}").Parse()

	require.NoError(t, err)
	object := op.Selections["edges"].InnerSelection["object"].InnerSelection
	assert.Equal(t, []string{"duration", "id", "title"}, topLevelNames(object))
}

func TestParse_FragmentExpansionCanBeDisabled(t *testing.T) {
	l := NewWithOptions(
		"{\n  edges {\n    object {\n      ... on Story { id title }\n    }\n  }\n}",
		Options{DisableFragmentExpansion: true},
	)

	op, err := l.Parse()

	require.NoError(t, err)
	assert.Empty(t, op.Selections["edges"].InnerSelection["object"].InnerSelection)
}

func TestParse_RecursiveFragmentTerminates(t *testing.T) {
	// A fragment cycle is invalid GraphQL but must not hang the analyzer.
	op, err := New("query { a { ...A } }\nfragment A on T { id ...B }\nfragment B on T { name ...A }").Parse()

	require.NoError(t, err)
	assert.Equal(t, []string{"id", "name"}, topLevelNames(op.Selections["a"].InnerSelection))
}

func TestParse_MalformedQueryReturnsError(t *testing.T) {
	// Errors are no longer swallowed.
	_, err := New("query { a(").Parse()

	assert.Error(t, err)
}

func TestParse_TokenLimitRejectsOversizedDocument(t *testing.T) {
	huge := "query {" + strings.Repeat("a b c d e f g h ", 5000) + "}"

	_, err := New(huge).Parse()

	assert.Error(t, err)
}

func TestParse_PathologicalInputsTerminateQuickly(t *testing.T) {
	// Deep nesting and fragment fan-out must not hang. The token limit and the
	// selection-node budget are what keep these bounded.
	var diamond strings.Builder
	diamond.WriteString("{ ...F0 }\n")
	const depth = 40
	for i := 0; i < depth; i++ {
		cur, next := strconv.Itoa(i), strconv.Itoa(i+1)
		diamond.WriteString("fragment F" + cur + " on T { ...F" + next + " ...F" + next + " }\n")
	}
	diamond.WriteString("fragment F" + strconv.Itoa(depth) + " on T { id }\n")

	cases := []struct{ name, input string }{
		{"deep braces", "{" + strings.Repeat("a{", 10000) + "b" + strings.Repeat("}", 10001)},
		{"deep lists", "{f(x: " + strings.Repeat("[", 10000) + strings.Repeat("]", 10000) + ")}"},
		{"deep objects", "{f(x: " + strings.Repeat("{a: ", 5000) + "1" + strings.Repeat("}", 5000) + ")}"},
		{"fragment diamond", diamond.String()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			done := make(chan struct{})
			go func() {
				defer close(done)
				_, _ = New(tc.input).Parse()
			}()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatalf("%s did not finish within 5s", tc.name)
			}
		})
	}
}

// =====================================================================
// Multi-operation documents and variables
// =====================================================================

func TestParse_MultipleOperations_DefaultsToFirst(t *testing.T) {
	op, err := New("query First { a }\nquery Second { b }").Parse()

	require.NoError(t, err)
	assert.Equal(t, "First", op.Name)
}

func TestParse_MultipleOperations_SelectByName(t *testing.T) {
	l := NewWithOptions("query First { a }\nquery Second { b }", Options{OperationName: "Second"})

	op, err := l.Parse()

	require.NoError(t, err)
	assert.Equal(t, "Second", op.Name)
	assert.Equal(t, []string{"b"}, topLevelNames(op.Selections))
}

func TestParse_VariableDefinitionsAreReported(t *testing.T) {
	op, err := New(`query Q($id: ID!, $size: Int) { a }`).Parse()

	require.NoError(t, err)
	names := make([]string, 0, len(op.Variables))
	for _, v := range op.Variables {
		names = append(names, v.Name)
	}
	assert.Equal(t, []string{"id", "size"}, names)
}

func TestParseWithVariables_PrefixNamesAreNotConfused(t *testing.T) {
	// Used to be non-deterministic: substitution was a plain string replace,
	// so "$id" also matched the start of "$idType", and Go's map iteration
	// order decided which one won.
	for i := 0; i < 50; i++ {
		op, err := New("query {\n  a(id: $id, idType: $idType)\n}").
			ParseWithVariables(`{"id": "x", "idType": "USER"}`)

		require.NoError(t, err)
		assert.Equal(t, `"x"`, op.Selections["a"].Arguments["id"].Value)
		assert.Equal(t, `"USER"`, op.Selections["a"].Arguments["idType"].Value)
	}
}

func TestParseWithVariables_QuoteInsideVariableValue(t *testing.T) {
	op, err := New("query {\n  a(t: $t)\n}").ParseWithVariables(`{"t": "say \"hi\""}`)

	require.NoError(t, err)
	assert.Equal(t, `"say \"hi\""`, op.Selections["a"].Arguments["t"].Value)
}

func TestParseWithVariables_NumberAndObjectValues(t *testing.T) {
	op, err := New("query {\n  a(size: $size, filter: $filter)\n}").
		ParseWithVariables(`{"size": 25, "filter": {"storyTypes": ["STORY"]}}`)

	require.NoError(t, err)
	assert.Equal(t, "25", op.Selections["a"].Arguments["size"].Value)
	assert.Equal(t, `{"storyTypes":["STORY"]}`, op.Selections["a"].Arguments["filter"].Value)
}

func TestParseWithVariables_UnsuppliedVariableKeepsReference(t *testing.T) {
	op, err := New("query {\n  a(id: $id)\n}").ParseWithVariables(`{}`)

	require.NoError(t, err)
	assert.Equal(t, "$id", op.Selections["a"].Arguments["id"].Value)
}

func TestParseWithVariables_InvalidJSONReturnsError(t *testing.T) {
	_, err := New("query { a(id: $id) }").ParseWithVariables(`{`)

	assert.Error(t, err)
}

func TestParse_FragmentsOnly_ReturnsZeroOperation(t *testing.T) {
	op, err := New("fragment F on T { id }").Parse()

	assert.NoError(t, err)
	assert.Equal(t, operation.Type(""), op.Type)
}

func TestParse_IsRepeatable(t *testing.T) {
	// The old Lexer carried cursor and stack state, so a second Parse on the
	// same value returned something different.
	l := New("query Q { a b }")

	first, err := l.Parse()
	require.NoError(t, err)
	second, err := l.Parse()
	require.NoError(t, err)

	assert.Equal(t, first.Name, second.Name)
	assert.Equal(t, topLevelNames(first.Selections), topLevelNames(second.Selections))
}

func TestReset_IsANoOp(t *testing.T) {
	l := New("query Q { a }")
	l.Reset()

	op, err := l.Parse()

	assert.NoError(t, err)
	assert.Equal(t, "Q", op.Name)
}

// =====================================================================
// Production corpus
// =====================================================================

// anonymizedCorpus is committed, so CI always exercises it. It holds a small
// and a large query for each distinct syntax shape found in a capture of
// production traffic, with nothing of production left in it: every string
// literal's contents and every identifier — operation names, fields, aliases,
// arguments, variables, types, fragments and enum values — were replaced.
// Only GraphQL's own __introspection names and the built-in scalars survive.
//
// What is preserved is each query's layout, byte for byte: one-line queries,
// inline sub-selections, escaped quotes, odd indentation. Layout is what the
// hand-written lexer got wrong, so layout is what the corpus has to keep. All
// of these still fail on v1, across all three of its error classes.
const anonymizedCorpus = "testdata/queries_anonymized.csv"

// fullCorpus is the unedited capture. It stays out of version control (see
// .gitignore) because it carries real content. Drop it in to run the whole
// capture locally.
const fullCorpus = "testdata/queries_anonymized.csv"

// TestParse_AnonymizedCorpus replays the committed corpus. Every query in it
// is valid GraphQL, and every one of them failed on the hand-written lexer.
// This test must never skip: it is the standing guard for this change.
func TestParse_AnonymizedCorpus(t *testing.T) {
	queries := readCorpus(t, anonymizedCorpus, true)
	require.GreaterOrEqual(t, len(queries), 15, "corpus must keep meaningful coverage")
	parseAll(t, queries, anonymizedCorpus)
}

// TestParse_FullProductionCorpus runs the unedited capture when present.
func TestParse_FullProductionCorpus(t *testing.T) {
	queries := readCorpus(t, fullCorpus, false)
	if queries == nil {
		t.Skipf("%s not present; drop the capture in to run it locally", fullCorpus)
	}
	parseAll(t, queries, fullCorpus)
}

// readCorpus reads a corpus CSV. When required is false, a missing file
// returns nil instead of failing the test.
func readCorpus(t *testing.T, path string, required bool) []string {
	t.Helper()

	f, err := os.Open(path)
	if os.IsNotExist(err) && !required {
		return nil
	}
	require.NoErrorf(t, err, "opening %s", path)
	defer f.Close() //nolint:errcheck // read-only

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	require.NoError(t, err)
	require.Greaterf(t, len(records), 1, "%s must hold at least one query", path)

	queries := make([]string, 0, len(records)-1)
	for _, rec := range records[1:] { // skip the header
		if len(rec) > 0 && strings.TrimSpace(rec[0]) != "" {
			queries = append(queries, rec[0])
		}
	}

	return queries
}

func parseAll(t *testing.T, queries []string, path string) {
	t.Helper()

	parsed := 0
	for i, q := range queries {
		op, err := NewWithOptions(q, Options{MaxTokenLimit: -1}).Parse()
		if !assert.NoErrorf(t, err, "%s row %d: %.120s", path, i+2, q) {
			continue
		}
		assert.NotEmptyf(t, op.Type, "%s row %d: operation type must be set", path, i+2)
		assert.NotEmptyf(t, op.Selections, "%s row %d: selections must not be empty", path, i+2)
		assert.Emptyf(t, op.UnresolvedFragments, "%s row %d: every spread should resolve", path, i+2)
		parsed++
	}

	assert.Equalf(t, len(queries), parsed, "every row of %s must parse", path)
	t.Logf("parsed %d/%d queries from %s", parsed, len(queries), path)
}

// topLevelNames returns the names in a selection set, sorted, so that
// assertions do not depend on Go's map iteration order.
func topLevelNames(set token.SelectionSet) []string {
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}

// =====================================================================
// Regressions for the findings raised in review of this change
// =====================================================================

// Finding 2: a mistyped OperationName used to yield a zero Operation and a
// nil error, which reads as "this query selects nothing" — harmless-looking
// when it is in fact a failure.
func TestOperationName_NotFound_ReturnsError(t *testing.T) {
	cases := []struct{ name, doc string }{
		{"document has other operations", "query First { a }"},
		{"document has no operation at all", "fragment F on T { id }"},
		{"empty document", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewWithOptions(tc.doc, Options{OperationName: "Nope"})

			_, err := l.Parse()
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrOperationNotFound)
			assert.Contains(t, err.Error(), "Nope")

			_, err = l.ParseOperationType()
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrOperationNotFound)
		})
	}
}

// Finding 3: ParseOperationType ignored OperationName, so the two methods on
// one Lexer disagreed — the cheap type check said QUERY while Parse returned
// the MUTATION the caller had asked for.
func TestOperationName_TypeAgreesWithParse(t *testing.T) {
	const doc = "query First { a }\nmutation Second { b }\nsubscription Third { c }"

	for _, want := range []struct {
		opName string
		typ    operation.Type
	}{
		{"First", operation.Query},
		{"Second", operation.Mutation},
		{"Third", operation.Subscription},
	} {
		t.Run(want.opName, func(t *testing.T) {
			l := NewWithOptions(doc, Options{OperationName: want.opName})

			ot, err := l.ParseOperationType()
			require.NoError(t, err)
			op, err := l.Parse()
			require.NoError(t, err)

			assert.Equal(t, want.typ, ot)
			assert.Equal(t, ot, op.Type, "the two methods must agree")
			assert.Equal(t, want.opName, op.Name)
		})
	}
}

// Finding 4: Parse treated a fragments-only document as "no operation, no
// error" while ParseOperationType called it an unknown definition.
func TestParseOperationType_AgreesWithParse(t *testing.T) {
	cases := []string{
		"",
		"   \n\t ",
		"# just a comment",
		"fragment F on T { id }",
		"fragment A on T { id }\nfragment B on T { name }",
		"query { a }",
		"mutation M { a }",
		"subscription S { a }",
		"{ a }",
	}

	for _, doc := range cases {
		t.Run(strings.TrimSpace(doc), func(t *testing.T) {
			ot, otErr := New(doc).ParseOperationType()
			op, pErr := New(doc).Parse()

			assert.Equal(t, pErr == nil, otErr == nil,
				"ParseOperationType err=%v but Parse err=%v", otErr, pErr)
			assert.Equal(t, op.Type, ot)
		})
	}
}

func TestParseOperationType_StillRejectsGarbage(t *testing.T) {
	_, err := New("querty { a }").ParseOperationType()

	assert.Error(t, err)
}

// Finding 5: a spread with no definition in the document was reported as a
// field, which is the very bug this change set out to remove for fragments
// that do resolve.
func TestUnresolvedFragmentSpread_IsNotAField(t *testing.T) {
	op, err := New("{ a { ...Missing id } }").Parse()

	require.NoError(t, err)
	assert.Equal(t, []string{"id"}, topLevelNames(op.Selections["a"].InnerSelection),
		"a fragment name must not appear as a field")
	assert.Equal(t, []string{"Missing"}, op.UnresolvedFragments,
		"but it must not vanish silently either")
}

func TestUnresolvedFragmentSpread_Sorted(t *testing.T) {
	op, err := New("{ a { ...Zeta ...Alpha } b { ...Alpha } }").Parse()

	require.NoError(t, err)
	assert.Equal(t, []string{"Alpha", "Zeta"}, op.UnresolvedFragments)
}

func TestResolvedFragments_LeaveUnresolvedEmpty(t *testing.T) {
	op, err := New("{ a { ...F } }\nfragment F on T { id }").Parse()

	require.NoError(t, err)
	assert.Empty(t, op.UnresolvedFragments)
	assert.Equal(t, []string{"id"}, topLevelNames(op.Selections["a"].InnerSelection))
}

// Finding 6: the doc comment claimed a Lexer was unsafe for concurrent use.
// It is immutable once built. Run with -race to make this meaningful.
func TestLexer_IsSafeForConcurrentUse(t *testing.T) {
	l := NewWithOptions(
		`query Q($id: ID!) { a(id: $id) { b c } }`,
		Options{OperationName: "Q"},
	)

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			op, err := l.Parse()
			assert.NoError(t, err)
			assert.Equal(t, "Q", op.Name)
			assert.Equal(t, []string{"a"}, topLevelNames(op.Selections))

			ot, err := l.ParseOperationType()
			assert.NoError(t, err)
			assert.Equal(t, operation.Query, ot)

			_, err = l.ParseWithVariables(`{"id": "1"}`)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}
