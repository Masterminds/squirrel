package squirrel

// Interface implementation, which allows to serialize sql queries.
type Serializer interface {
	Select(data selectData) (sqlStr string, args []interface{}, err error)
	Update(data updateData) (sqlStr string, args []interface{}, err error)
	Delete(data deleteData) (sqlStr string, args []interface{}, err error)
	Insert(data insertData) (sqlStr string, args []interface{}, err error)
	Case(data caseData) (sqlStr string, args []interface{}, err error)

	EQ(eq Eq, useNotOpr bool) (sql string, args []interface{}, err error)
	LT(lt Lt, opposite, orEq bool) (sql string, args []interface{}, err error)
}