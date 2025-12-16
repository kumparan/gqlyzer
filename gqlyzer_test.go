package gqlyzer

import (
	"testing"

	"github.com/kumparan/gqlyzer/token/operation"
	"github.com/stretchr/testify/assert"
)

func TestParseWithVariable(t *testing.T) {
	l := Lexer{input: `query SomeOperation {
			SomeQuery(id: $id) {
				subQuery
			}
		}`}
	l.Reset()
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
		l := Lexer{input: `{
	  IniQuerySatu(
	    id: "aya" object: USER
	  )
	  IniQueryDua(
	    id: "aya" object: USER
	  )
	}`}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("graphql query without variable", func(t *testing.T) {
		l := Lexer{input: `query iniOperationName {
	  IniQuerySatu(
	    id: "aya" object: USER
	  )
	  IniQueryDua(
	    id: "aya" object: USER
	  )
	}`}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("graphql query with variable", func(t *testing.T) {
		l := Lexer{input: `query iniOperationName(
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
	}`}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("json without opName, with var", func(t *testing.T) {
		l := Lexer{input: "query ( $objectID: ID! $userID: ID! $objectType: ObjectType!\t) {\n IniQuerySatu( objectID: $objectID\n objectType: $objectType )\n IniQueryDua( userID: $userID\n objectType: $objectType )\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("json with opName, with var", func(t *testing.T) {
		l := Lexer{input: "query iniOperationName( $objectID: ID! $userID: ID! $objectType: ObjectType!) {\n IniQuerySatu( objectID: $objectID\n objectType: $objectType )\n IniQueryDua( userID: $userID\n objectType: $objectType )\n}\n"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("json with opName, without var", func(t *testing.T) {
		l := Lexer{input: "query iniOperationName {\n IniQuerySatu(id: \"19\", object: USER)}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
	})

	t.Run("json with opName, without var, with fragments", func(t *testing.T) {
		l := Lexer{input: "query iniOperationName {\n  IniQuerySatu {\n    ...FragmentExample\n    __typename\n  }\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
	})

	t.Run("json without opName, without var", func(t *testing.T) {
		l := Lexer{input: "query {\n IniQuerySatu(userID: \"19\", objectType: USER )\n IniQueryDua(userID: \"19\", objectType: USER )\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "IniQueryDua", s.Selections["IniQueryDua"].Name)
	})

	t.Run("json with opName, with var, with alias", func(t *testing.T) {
		l := Lexer{input: "query iniOperationName($id: ID!) {\n  iniQueryAliasSatu: IniQuerySatu(id: $id) {\n    ...ItemDetails\n  }\n}\n\nfragment ItemDetails on Item {\n  ...BasicInfo\n  price\n}\n\nfragment BasicInfo on Item {\n  id\n  name\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "iniQueryAliasSatu", s.Selections["IniQuerySatu"].Alias)
	})

	t.Run("json with opName, without var, with alias", func(t *testing.T) {
		l := Lexer{input: "query iniOperationName {\n  iniQueryAliasSatu: IniQuerySatu {\n    ...FragmentExample\n    __typename\n  }\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "iniQueryAliasSatu", s.Selections["IniQuerySatu"].Alias)
	})

	t.Run("json with opname, with var, with alias, with object value arg", func(t *testing.T) {
		l := Lexer{input: "query iniOperationName($objectID: ID! $userID: ID! $objectType: ObjectType!) {\n\t\tiniQueryAliasSatu: IniQuerySatu(objectID: $objectID, userID: $userID, objectType: $objectType, filter: {\n\t\t\tiniObjectValueArgument: false\n\t\t\tiniJuga: $iniJuga\n\t\t}) {\n\t\t\tedges {\n\t\t\t\tid\n\t\t\t\ttitle\n\t\t\t\tpublisher {\n\t\t\t\t\tslug\n\t\t\t\t}\n\t\t\t\tauthor {\n\t\t\t\t\tusername\n\t\t\t\t}\n\t\t\t\tvideo {\n\t\t\t\t\tid\n\t\t\t\t\tduration\n\t\t\t\t\torientation\n\t\t\t\t\tposterMedia {\n\t\t\t\t\t\texternalURL\n\t\t\t\t\t}\n\t\t\t\t}\n\t\t\t\tcaption {\n\t\t\t\t\tdocument\n\t\t\t\t}\n\t\t\t\tcreatedAt\n\t\t\t}\n\t\t}\n\t}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
		assert.Equal(t, "iniQueryAliasSatu", s.Selections["IniQuerySatu"].Alias)
	})

	t.Run("json with opName, without variable, w/o alias, with object value arg no line feed", func(t *testing.T) {
		l := Lexer{input: "query iniOperationName {\n  IniQuerySatu(\n    query: \"\"\n    size: 1\n    cursor: \"1\"\n    cursorType: PAGE\n    filters: {status: NEED_REVIEW}\n    sortType: STATUS_ASC_AND_UPDATED_AT_DESC\n  ) {\n    cursorInfo {\n      count\n      __typename\n    }\n    __typename\n  }\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "iniOperationName", s.Name)
		assert.Equal(t, "IniQuerySatu", s.Selections["IniQuerySatu"].Name)
	})

	t.Run("json with opName, with variable, w/o alias, with fragment", func(t *testing.T) {
		l := Lexer{input: "mutation AddObjectToProfileClassification($objectID: ID!, $objectType: ProfileClassificationObjectType!, $profileClassificationID: ID!) {\n\nAddObjectToProfileClassification(objectID: $objectID, objectType: $objectType, profileClassificationID: $profileClassificationID){\n\n... on User{\n\n...User\n\n}\n\n... on Publisher{\n\n...Publisher\n\n}\n\n}\n\n}\n\nfragment User on User {\n\n__typename\n\nid\n\nname\n\nusername\n\naboutMe\n\nemail\n\nstatus\n\nphone\n\nemailVerified\n\nphoneVerified\n\nprofilePictureMedia {\n\n...Media\n\n}\n\ncoverPictureMedia {\n\n...Media\n\n}\n\ngender\n\nuserStatus: status\n\nbirthDate\n\nisRecommended\n\ncreatedAt\n\nupdatedAt\n\ndeletedAt\n\naboutMe\n\nisVerified\n\nwebsiteURL\n\nisVerified\n\nemailVerified\n\nwebsiteURL\n\nrole{\n\nid\n\nname\n\nslug\n\n}\n\nlastUpdatedBy{\n\nid\n\n}\n\nmetaTitle\n\nmetaDescription\n\nmetaKeyword\n\nemails{\n\nemail\n\nverifiedAt\n\ncreatedAt\n\n}\n\nisPasswordSet\n\nauthorizedChannel{\n\nisAuthorized\n\nchannel{\n\nid\n\nname\n\nslug\n\nmeta_title\n\nmeta_description\n\nmeta_keywords\n\n}\n\n}\n\nuserTermsAndConditionsAgreement {\n\nagreedAt\n\nid\n\nstatus\n\ntermsAndConditions {\n\ncreatedAt\n\nid\n\nupdatedAt\n\nversion\n\n}\n\n}\n\n}\n\nfragment Media on Media {\n\nid\n\ntitle\n\ndescription\n\npublicID\n\nexternalURL\n\nawsS3Key\n\nheight\n\nwidth\n\nlocationName\n\nlocationLat\n\nlocationLon\n\nmediaType\n\nmediaSourceID\n\nphotographer\n\neventDate\n\nlastUpdatedBy{\n\nid\n\nname\n\n}\n\nisArchived\n\ncreatedBy{\n\nid\n\nname\n\n}\n\ncreatedAt\n\nlastUpdatedAt\n\ntopics{\n\nid\n\n}\n\nmediaSource{\n\nid\n\nname\n\ncreatedAt\n\nlastUpdatedAt\n\ncreatedBy{\n\nid\n\nname\n\n}\n\nlastUpdatedBy{\n\nid\n\nname\n\n}\n\n}\n\n}\n\nfragment Publisher on Publisher {\n\n__typename\n\nid\n\nname\n\nslug\n\ndescription\n\nwebsite\n\nmetaTitle\n\nmetaKeywords\n\nmetaDescription\n\nisVerified\n\nisActive\n\nisPremium\n\ncoverMedia {\n\n...SimpleMedia\n\n}\n\navatarMedia {\n\n...SimpleMedia\n\n}\n\norganisation{\n\n...Organisation\n\n}\n\nauthorizedRSSConsumers{\n\nid\n\n}\n\nauthorizedChannel{\n\nchannel{\n\nid\n\n}\n\nisAuthorized\n\n}\n\npublisherGroupID\n\nisAutoMemberByDomain\n\ndomains\n\nenableGeneralPushNotificationForMember\n\nenableSegmentedPushNotificationForMember\n\n}\n\nfragment SimpleMedia on Media {\n\nid\n\ntitle\n\nlastUpdatedBy{\n\nid\n\n}\n\nisArchived\n\ncreatedBy{\n\nid\n\n}\n\ncreatedAt\n\nlastUpdatedAt\n\ntopics{\n\nid\n\n}\n\nmediaSource{\n\nid\n\ncreatedBy{\n\nid\n\n}\n\nlastUpdatedBy{\n\nid\n\n}\n\n}\n\n}\n\nfragment Organisation on Organisation {\n\nid\n\nname\n\nslug\n\ndescription\n\norganisationType\n\nwebsite\n\nisActive\n\ncoverMedia{\n\n...SimpleMedia\n\n}\n\navatarMedia{\n\n...SimpleMedia\n\n}\n\naddress\n\nphone1\n\nphone2\n\nemail\n\nmetaTitle\n\nmetaDescription\n\nmetaKeywords\n\nownedBy{\n\n...SimpleUser\n\n}\n\ncreatedBy{\n\n...SimpleUser\n\n}\n\n}\n\nfragment SimpleUser on User {\n\n__typename\n\nid\n\nname\n\nusername\n\nrole{\n\nid\n\n}\n\nauthorizedChannel{\n\nisAuthorized\n\nchannel{\n\nid\n\n}\n\n}\n\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Mutation, s.Type)
		assert.Equal(t, "AddObjectToProfileClassification", s.Name)
		assert.Equal(t, "AddObjectToProfileClassification", s.Selections["AddObjectToProfileClassification"].Name)
	})

	t.Run("json with text", func(t *testing.T) {
		l := Lexer{input: "mutation {\n  ReviseTopicSummaries(\n    linkedSummaryID: \"12345678\"\n    synthesisVoiceID: 2\n    reviseInput: [\n      {\n        summaryID: \"12345678\"\n        revisedSummary: \"COK Suzuki Fronx adalah mobil sub-compact SUV yang dirilis dengan harga mulai dari Rp 242,2 juta hingga Rp 316,3 juta. asda asda sdas\"\n      }\n    ]\n  )\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Mutation, s.Type)
		assert.Equal(t, "ReviseTopicSummaries", s.Selections["ReviseTopicSummaries"].Name)
	})

	t.Run("json with text 2", func(t *testing.T) {
		l := Lexer{input: "mutation {\n\tAnalyzeTypo(texts: [\n    \"GuluGuluGleg Gleg Gleg Khhrkkkrrrhrhhhrkkk\",\n    \"Lorem ipsum dolor sit amet, elit\",\n    \"roin maximus lectus ut turpis semper, vel blandit est accumsan.\",\n    \"Quisque faucibus, dui eu suscipit condimentum, sapien ante tincidunt ipsum, vitae aliquam elit odio quis arcu.\",\n    \"Donec aliquet tristique elit ut euismod\",\n    \"Proin ut urna eget mi euismod auctor.\",\n    \"Quisque faucibus, dui eu suscipit condimentum, sapien ante tincidunt ipsum, vitae aliquam elit odio quis arcu.\",\n\t]) {\n\t\ttypos{\n\t\t\toffset\n\t\t\ttype\n\t\t\ttoken\n\t\t\tsuggestions {\n\t\t\t\ttoken\n\t\t\t\tscore\n\t\t\t}\n\t\t}\n\t\t\n\t}\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Mutation, s.Type)
		assert.Equal(t, "AnalyzeTypo", s.Selections["AnalyzeTypo"].Name)
	})

	t.Run("json input array", func(t *testing.T) {
		l := Lexer{input: "mutation {\n\tReviseTopicSummaries(\n\t\treviseInput: [\n\t\t\t{\n\t\t\tsummaryID: \"1234567890\"\n\t\t\trevisedSummary: \"Satpol PP Kabupaten Penajam Paser Utara menangkap 64 PSK di wilayah IKN sepanjang tahun ini. Mereka yang terjaring berasal dari berbagai kota seperti Samarinda, Balikpapan, Bandung, Makassar, dan Yogyakarta. \\n \\n Para PSK tersebut beroperasi secara mandiri, tanpa difasilitasi oleh muncikari. Oleh karena itu, mereka tidak dapat dikenakan pidana, melainkan hanya mendapatkan sanksi pengusiran dari Penajam Paser Utara.\"\n\t\t\t},\n\t\t\t{\n\t\t\tsummaryID: \"1234567890\"\n\t\t\trevisedSummary: \"* Terbaru! Suzuki Fronx meluncurkan fitur Advanced Driving Assistant System (ADAS) yang dirancang untuk membantu pengemudi mengurangi keletihan dan meningkatkan keselamatan berkendara.\\n* Fitur-fitur ADAS di Suzuki Fronx meliputi Dual Sensor Brake Support II, Adaptive Cruise Control, Lane Keep Assist, Lane Departure Warning, Lane Departure Prevention, Vehicle Swaying Warning, Blind Spot Monitor, Rear Cross Traffic Alert, dan High Beam Assist.\\n* Sistem ini memanfaatkan modul kamera dan sensor radar untuk memancarkan gelombang radio untuk mengukur jarak dan kecepatan objek di depan dan belakang.\\n* Suzuki Fronx hadir sebagai pilihan baru di segmen SUV sub-compact crossover dengan panjang dimensi 4 meter dan tersedia dalam varian SGX A/T SHVS, GX A/T SHVS, GX M/T SHVS, GL A/T, dan GL M/T.\\n* Suzuki Fronx berhasil mengimpor 3.990 unit mobil ke Jepang pada bulan April, lebih tinggi dibandingkan Mercedes-Benz dan BMW, didorong oleh ledakan permintaan terhadap Jimny Nomade versi lima pintu.\"\n\t\t\t},\n\t\t]\n\t\tlinkedSummaryID: \"1765524525286736337\"\n\t\tsynthesisVoiceID: \"1\"\n\t)\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.NoError(t, err)
		assert.Equal(t, operation.Mutation, s.Type)
		assert.Equal(t, "ReviseTopicSummaries", s.Selections["ReviseTopicSummaries"].Name)
	})

	t.Run("json introspection query", func(t *testing.T) {
		l := Lexer{input: "\n    query IntrospectionQuery {\n      __schema {\n        \n        queryType { name kind }\n        mutationType { name kind }\n        subscriptionType { name kind }\n        types {\n          ...FullType\n        }\n        directives {\n          name\n          description\n          \n          locations\n          args {\n            ...InputValue\n          }\n        }\n      }\n    }\n\n    "}
		l.Reset()

		s, err := l.Parse()

		assert.Error(t, err) // TODO: should be no error. have not yet handle subquery with space separator "{ name kind }"
		assert.Equal(t, operation.Query, s.Type)
		assert.Equal(t, "IntrospectionQuery", s.Name)
		assert.Equal(t, "__schema", s.Selections["__schema"].Name)
	})

	t.Run("json query like user input", func(t *testing.T) {
		l := Lexer{input: "mutation {\n  CreateDraftStoryV2(\n    draft: {\n\t\t\tauthorID: \"1234567890\", \n\t\t\tpublisherID: \"\", \n\t\t\tchannelID: \"3\", \n\t\t\ttitle: \"ICOK Suzuki Fronx adalah mobil\", \n\t\t\tsource: UGC, \n\t\t\tleadText: \"dummy leadtext\", \n\t\t\tcontent: \"{\\\"object\\\":\\\"value\\\",\\\"document\\\":{\\\"object\\\":\\\"document\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"heading-large\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"IHSG Dibuka Menguat, Rupiah Melemah, Bursa Asia Bergerak Variatif\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Indeks Harga Saham Gabungan (IHSG) mengawali perdagangan hari ini dengan menunjukkan penguatan, mencerminkan sentimen positif di awal sesi. Namun, di pasar valuta asing, nilai tukar rupiah terhadap dolar Amerika Serikat (AS) terpantau melemah. Sementara itu, bursa saham-saham utama di Asia menampilkan pergerakan yang beragam, dengan beberapa indeks dibuka positif dan lainnya menunjukkan koreksi di sesi pertama.\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"heading-medium\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"IHSG\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Pada pembukaan perdagangan tanggal \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"11 Desember 2025\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"09:00:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", Indeks Harga Saham Gabungan (IHSG) berhasil dibuka di level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"8764.09\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\". Kinerja positif ini ditandai dengan kenaikan sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"0.73%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" dari posisi penutupan sebelumnya, memberikan sinyal optimisme bagi para investor di awal sesi perdagangan.\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"heading-medium\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Nilai Tukar Rupiah\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Kondisi berbeda terlihat di pasar valuta asing. Data terkini pada pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"09:50:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" menunjukkan bahwa nilai tukar mata uang rupiah terhadap dolar AS berada pada level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"16683\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\". Angka ini mengindikasikan adanya depresiasi atau pelemahan rupiah terhadap mata uang Negeri Paman Sam tersebut, yang mungkin menjadi perhatian bagi eksportir dan importir serta sektor keuangan.\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"heading-medium\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Bursa Saham Asia\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Berikut adalah kinerja indeks saham utama di Asia pada awal perdagangan hari ini:\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Nikkei 225\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" (Jepang):\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   Dibuka pada \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"11 Desember 2025\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"07:00:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" di level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"50818.39\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", naik sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"0.43%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\".\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   Pada penutupan sesi 1 pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"09:35:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", indeks ini bergerak terkoreksi ke level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"50308.89\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", turun sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"-0.58%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\".\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Hang Seng Index\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" (Hong Kong):\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   Dibuka pada \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"11 Desember 2025\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"08:30:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" di level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"25710.61\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", menguat sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"0.66%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\".\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Shanghai Composite\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" (Tiongkok):\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   Dibuka pada \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"11 Desember 2025\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"08:30:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" di level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"3904.96\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", naik tipis sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"0.11%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\".\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"STI (Straits Times Index)\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" (Singapura):\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"*   Dibuka pada \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"11 Desember 2025\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" pukul \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"08:00:00 WIB\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\" di level \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"4516.34\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\", menguat sebesar \\\",\\\"marks\\\":[]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"0.23%\\\",\\\"marks\\\":[{\\\"object\\\":\\\"mark\\\",\\\"type\\\":\\\"bold\\\"}]},{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\".\\\",\\\"marks\\\":[]}]}]},{\\\"object\\\":\\\"block\\\",\\\"type\\\":\\\"paragraph\\\",\\\"data\\\":{},\\\"nodes\\\":[{\\\"object\\\":\\\"text\\\",\\\"leaves\\\":[{\\\"object\\\":\\\"leaf\\\",\\\"text\\\":\\\"Secara keseluruhan, bursa saham Asia menunjukkan pergerakan yang variatif. Meskipun Hang Seng, Shanghai Composite, dan STI dibuka dengan penguatan, Nikkei 225 yang sebelumnya dibuka positif harus mengalami koreksi pada penutupan sesi pertamanya. Hal ini mencerminkan sentimen pasar yang beragam di kawasan Asia pagi ini, dengan beberapa pasar masih menjaga momentum positif sementara yang lain menghadapi tekanan jual.\\\",\\\"marks\\\":[]}]}]}]}}\", \n\t\t\tdocumentType: SLATEJS, \n\t\t\treporterIDs: [], \n\t\t\tleadMediaIDs: [], \n\t\t\ttopicIDs: [], \n\t\t\teditorIDs: [], \n\t\t\taddOns: [], \n\t\t\tattributes: {}\n\t\t}) \n\t{\n    id\n    title\n    slug\n  }\n}"}
		l.Reset()

		s, err := l.Parse()

		assert.Error(t, err) // TODO: handle query-like input
		assert.Equal(t, operation.Mutation, s.Type)
		assert.Equal(t, "CreateDraftStoryV2", s.Selections["CreateDraftStoryV2"].Name)
	})
}
