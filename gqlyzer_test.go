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
}
