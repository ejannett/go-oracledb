/*
** Copyright (c) 2026 Oracle and/or its affiliates.
**
** The Universal Permissive License (UPL), Version 1.0
**
** Subject to the condition set forth below, permission is hereby granted to any
** person obtaining a copy of this software, associated documentation and/or data
** (collectively the "Software"), free of charge and under any and all copyright
** rights in the Software, and any and all patent rights owned or freely
** licensable by each licensor hereunder covering either (i) the unmodified
** Software as contributed to or provided by such licensor, or (ii) the Larger
** Works (as defined below), to deal in both
**
** (a) the Software, and
** (b) any piece of software and/or hardware listed in the lrgrwrks.txt file if
** one is included with the Software (each a "Larger Work" to which the Software
** is contributed by such licensors),
**
** without restriction, including without limitation the rights to copy, create
** derivative works of, display, perform, and distribute the Software and make,
** use, sell, offer for sale, import, export, have made, and have sold the
** Software and the Larger Work(s), and to sublicense the foregoing rights on
** either these or other terms.
**
** This license is subject to the following condition:
** The above copyright notice and either this complete permission notice or at
** a minimum a reference to the UPL must be included in all copies or
** substantial portions of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
** IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
** FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
** AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
** LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
** OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
 */

package oracle

import (
	"fmt"
	"os"
	"testing"

	oracleTest "github.com/oracle/go-oracledb/v26/internal/tests"
)

func TestMain(m *testing.M) {
	err := oracleTest.InitConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "InitConfig failed: %v\n", err)
		os.Exit(1)
	}
	TestEnvironement = oracleTest.TestEnvironement
	TestingConfig = oracleTest.TestingConfig
	DefaultTestConfig = oracleTest.DefaultTestConfig
	os.Exit(m.Run())
}

var testCases = []oracleTest.CategorizedTestCase{
	{"TestDriver_ConfigurationWithConnectorBasic", "functional", true, TestDriver_ConfigurationWithConnectorBasic},
	{"TestDriver_ConfigurationWithConnectorWithEnvOverwrite", "functional", true, TestDriver_ConfigurationWithConnectorWithEnvOverwrite},
	{"TestDriver_ConfigurationWithConnectorWithFlagOverwrite", "functional", true, TestDriver_ConfigurationWithConnectorWithFlagOverwrite},
	{"TestConfiguration_AssignFromEmptyMap", "unitary", false, TestConfiguration_AssignFromEmptyMap},
	{"TestConfiguration_AssignFromMapUnknownKey", "unitary", false, TestConfiguration_AssignFromMapUnknownKey},
	{"TestConfiguration_AssignFromMap", "unitary", false, TestConfiguration_AssignFromMap},
	{"TestConfiguration_AssignFromMapValidatedIntString", "unitary", false, TestConfiguration_AssignFromMapValidatedIntString},
	{"TestConfiguration_AssignFromEnv", "unitary", true, TestConfiguration_AssignFromEnv},
	{"TestConfiguration_AssignFromEnvValidatedIntString", "unitary", true, TestConfiguration_AssignFromEnvValidatedIntString},
	{"TestConfiguration_AssignFromEmptyFlags", "unitary", false, TestConfiguration_AssignFromEmptyFlags},
	{"TestConfiguration_Clone", "unitary", false, TestConfiguration_Clone},
	{"TestConfiguration_DefaultClientLanguageIsLanguageTag", "unitary", false, TestConfiguration_DefaultClientLanguageIsLanguageTag},
	{"TestConfiguration_AssignFromMapClientLanguageTag", "unitary", false, TestConfiguration_AssignFromMapClientLanguageTag},
	{"TestConfiguration_AssignFromEnvClientLanguageTag", "unitary", true, TestConfiguration_AssignFromEnvClientLanguageTag},
	{"TestConfiguration_toNSConnectionParameters", "unitary", false, TestConfiguration_toNSConnectionParameters},
	{"TestConfiguration_InitLoggingWithConfigFileDestination", "unitary", false, TestConfiguration_InitLoggingWithConfigFileDestination},
	{"TestEnquoteLiteral", "unitary", false, TestEnquoteLiteral},
	{"TestEnquoteNCharLiteral", "unitary", false, TestEnquoteNCharLiteral},
	{"TestIsSimpleIdentifier", "unitary", false, TestIsSimpleIdentifier},
	{"TestEnquoteIdentifier", "unitary", false, TestEnquoteIdentifier},
	{"TestDriver_ConfigurationWithCredentialsWithDsnNegative", "unitary", false, TestDriver_ConfigurationWithCredentialsWithDsnNegative},
	{"TestDriver_ConfigurationLogging", "unitary", false, TestDriver_ConfigurationLogging},
	{"TestDriver_OpenConnectorUsesNSParamOverConfig", "unitary", false, TestDriver_OpenConnectorUsesNSParamOverConfig},
	{"TestDriver_Table_Create", "sanity", false, TestDriver_Table_Create},
	{"TestDriver_DropTable_DeniesAccess", "functional", false, TestDriver_DropTable_DeniesAccess},
	{"TestDriver_AlterSessionSetLanguage", "functional", false, TestDriver_AlterSessionSetLanguage},
	{"TestDriver_Table_Insert", "functional", false, TestDriver_Table_Insert},
	{"TestDriver_Insert_Select", "functional", false, TestDriver_Insert_Select},
	{"TestDriver_PLSQL_AnonymousBlock_Sanity", "functional", false, TestDriver_PLSQL_AnonymousBlock_Sanity},
	{"TestDriver_PLSQL_CreateInsertSelectDrop", "functional", false, TestDriver_PLSQL_CreateInsertSelectDrop},
	{"TestDriver_PLSQL_BreakCausedByTimeout", "functional", false, TestDriver_PLSQL_BreakCausedByTimeout},
	{"TestDriver_Select_BooleanTypes_23c", "functional", false, TestDriver_Select_BooleanTypes_23c},
	{"TestDriver_Select_CharacterTypes", "functional", false, TestDriver_Select_CharacterTypes},
	{"TestDriver_Select_DATE", "functional", false, TestDriver_Select_DATE},
	{"TestDriver_Select_TIMESTAMP", "functional", false, TestDriver_Select_TIMESTAMP},
	{"TestDriver_Select_TimestampWithTimeZone", "functional", false, TestDriver_Select_TimestampWithTimeZone},
	{"TestDriver_Select_TimestampWithLocalTimeZone", "functional", false, TestDriver_Select_TimestampWithLocalTimeZone},
	{"TestDriver_Select_Intervals", "functional", false, TestDriver_Select_Intervals},
	{"TestDriver_Select_NumericFloatTypes", "functional", false, TestDriver_Select_NumericFloatTypes},
	{"TestDriver_Select_Number_NoPrecisionScale", "functional", false, TestDriver_Select_Number_NoPrecisionScale},
	{"TestDriver_Select_NumberPrecision", "functional", false, TestDriver_Select_NumberPrecision},
	{"TestDriver_VarcharLargePayload", "functional", false, TestDriver_VarcharLargePayload},
	{"TestDriver_Select_Number_MaxPrecisionScale", "functional", false, TestDriver_Select_Number_MaxPrecisionScale},
	{"TestDriver_Select_Number_MaxPrecisionInteger", "functional", false, TestDriver_Select_Number_MaxPrecisionInteger},
	{"TestDriver_Select_Number_And_BinaryDouble", "functional", false, TestDriver_Select_Number_And_BinaryDouble},
	{"TestDriver_Table_Select", "functional", false, TestDriver_Table_Select},
	{"TestTimeoutConnectWithTransportConnectTimeout", "functional", false, TestTimeoutConnectWithTransportConnectTimeout},
	{"TestTimeoutConnectWithTransportConnectTimeoutAndContext", "functional", false, TestTimeoutConnectWithTransportConnectTimeoutAndContext},
	{"TestTimeoutConnectWithTransportConnectTimeoutAndContextGreater", "functional/cyclops", false, TestTimeoutConnectWithTransportConnectTimeoutAndContextGreater},
	{"TestTimeoutConnectWithRecvTimeout", "functional/cyclops", false, TestTimeoutConnectWithRecvTimeout},
	{"TestTimeoutConnectWithConnectTimeout", "functional/cyclops", false, TestTimeoutConnectWithConnectTimeout},
	{"TestTimeoutConnectWithRecvConnectTimeoutAndContext", "functional/cyclops", false, TestTimeoutConnectWithRecvConnectTimeoutAndContext},
	{"TestTimeoutConnectWithRecvConnectTimeoutAndContextGreater", "functional/cyclops", false, TestTimeoutConnectWithRecvConnectTimeoutAndContextGreater},
	{"TestTimeoutConnectTimeoutPrecedence1", "functional/cyclops", false, TestTimeoutConnectTimeoutPrecedence1},
	{"TestTimeoutConnectTimeoutPrecedence2", "functional/cyclops", false, TestTimeoutConnectTimeoutPrecedence2},
	{"TestTimeoutConnectTimeoutPrecedence3", "functional/cyclops", false, TestTimeoutConnectTimeoutPrecedence3},
	{"TestTimeoutConnectTimeoutPrecedence4", "functional", false, TestTimeoutConnectTimeoutPrecedence4},
	{"TestDriver_Functional_SelectDual", "sanity,functional", false, TestDriver_Functional_SelectDual},
	{"TestDriver_SimpleConnection", "sanity,functional", false, TestDriver_SimpleConnection},
	{"TestDriver_Authentication_TTIWRN", "functional", false, TestDriver_Authentication_TTIWRN},
	{"TestDriver_Authentication_OCIToken", "functional", false, TestDriver_Authentication_OCIToken},
	{"TestDriver_Authentication_OAuth", "functional", false, TestDriver_Authentication_OAuth},
	{"TestDriver_TCPS_Pipeline_SelectDual", "sanity,functional", false, TestDriver_TCPS_Pipeline_SelectDual},
	{"TestDriver_TCPS_Pipeline_InvalidCertDn", "sanity,functional", false, TestDriver_TCPS_Pipeline_InvalidCertDn},
	{"TestDriver_TCPS_Pipeline_InvalidWalletLocation", "sanity,functional", false, TestDriver_TCPS_Pipeline_InvalidWalletLocation},
	{"TestDriver_TCPS_Pipeline_DNMatchOff_AllowsInvalidDN", "sanity,functional", false, TestDriver_TCPS_Pipeline_DNMatchOff_AllowsInvalidDN},
	{"TestDriver_Prepared_Insert_Select_Ordinal", "sanity,functional", false, TestDriver_Prepared_Insert_Select_Ordinal},
	{"TestDriver_Prepared_Insert_Select_Named", "functional", false, TestDriver_Prepared_Insert_Select_Named},
	{"TestDriver_PLSQL_Prepared_Binds", "functional", false, TestDriver_PLSQL_Prepared_Binds},
	{"TestDriver_Select_BooleanTypes_23c_Prepared_Statement", "functional", false, TestDriver_Select_BooleanTypes_23c_Prepared_Statement},
	{"TestDriver_Select_BooleanTypes_23c_Prepared_Statement_Named", "functional", false, TestDriver_Select_BooleanTypes_23c_Prepared_Statement_Named},
	{"TestDriver_Select_BooleanTypes_19c", "functional", false, TestDriver_Select_BooleanTypes_19c},
	{"TestDriver_Select_BooleanTypes_19c_Prepared_Statement", "functional", false, TestDriver_Select_BooleanTypes_19c_Prepared_Statement},
	{"TestDriver_Select_BooleanTypes_19c_Prepared_Statement_Named", "functional", false, TestDriver_Select_BooleanTypes_19c_Prepared_Statement_Named},
	{"TestDriver_Select_CharacterTypes_Ordinal", "functional", false, TestDriver_Select_CharacterTypes_Ordinal},
	{"TestDriver_Select_CharacterTypes_Named", "functional", false, TestDriver_Select_CharacterTypes_Named},
	{"TestDriver_Select_DATE_Prepared_Named", "functional", false, TestDriver_Select_DATE_Prepared_Named},
	{"TestDriver_Select_TIMESTAMP_Prepared_Named", "functional", false, TestDriver_Select_TIMESTAMP_Prepared_Named},
	{"TestDriver_Select_TimestampWithTimeZone_Prepared_Named", "functional", false, TestDriver_Select_TimestampWithTimeZone_Prepared_Named},
	{"TestDriver_Select_TimestampWithLocalTimeZone_Prepared_Named", "functional", false, TestDriver_Select_TimestampWithLocalTimeZone_Prepared_Named},
	{"TestDriver_Select_TimestampWithTimeZone_Prepared_LoadLocation_And_Offset", "functional", false, TestDriver_Select_TimestampWithTimeZone_Prepared_LoadLocation_And_Offset},
	{"TestDriver_Prepared_InsertAndSelect_AllTypes", "functional", false, TestDriver_Prepared_InsertAndSelect_AllTypes},
	{"TestDriver_Prepared_InsertAndSelect_AllTypes_StrictNulls", "functional", false, TestDriver_Prepared_InsertAndSelect_AllTypes_StrictNulls},
	{"TestDriver_Prepared_InsertAndSelect_AllTypes_DefaultValuesForNulls", "functional", false, TestDriver_Prepared_InsertAndSelect_AllTypes_DefaultValuesForNulls},
	{"TestDriver_Prepared_Statement_Re_exec_With_Different_Arg_Counts_Negative", "functional", false, TestDriver_Prepared_Statement_Re_exec_With_Different_Arg_Counts_Negative},
	{"TestDriver_Exec_Query_cursor_leak", "robustness", false, TestDriver_Exec_Query_cursor_leak},
	{"TestDriver_Select_Query_cursor_leak", "robustness", false, TestDriver_Select_Query_cursor_leak},
	{"TestDriver_PreparedStatement_Query_cursor_leak", "robustness", false, TestDriver_PreparedStatement_Query_cursor_leak},
	{"TestQueryNonExistentTable_NegativeCase", "functional", false, TestQueryNonExistentTable_NegativeCase},
	{"TestPreparedStatementNonExistentTable_NegativeCase", "functional", false, TestPreparedStatementNonExistentTable_NegativeCase},
	{"TestSelectSpecificColumnsNonExistentTable_NegativeCase", "functional", false, TestSelectSpecificColumnsNonExistentTable_NegativeCase},
	{"TestCountQueryNonExistentTable_NegativeCase", "functional", false, TestCountQueryNonExistentTable_NegativeCase},
	{"TestJoinWithNonExistentTable_NegativeCase", "functional", false, TestJoinWithNonExistentTable_NegativeCase},
	{"TestSubqueryWithNonExistentTable_NegativeCase", "functional", false, TestSubqueryWithNonExistentTable_NegativeCase},
	{"TestDescribeNonExistentTable_NegativeCase", "functional", false, TestDescribeNonExistentTable_NegativeCase},
	{"TestInvalidTableNameSyntax_NegativeCase", "functional", false, TestInvalidTableNameSyntax_NegativeCase},
	{"TestQuerySystemTableWithoutPrivilege_NegativeCase", "functional", false, TestQuerySystemTableWithoutPrivilege_NegativeCase},
	{"TestQueryAccessibleTable_PositiveCase", "functional", false, TestQueryAccessibleTable_PositiveCase},
	{"TestQueryDictionaryViewWithoutPrivilege_NegativeCase", "functional", false, TestQueryDictionaryViewWithoutPrivilege_NegativeCase},
	{"TestInsertAndSelectSmallRAW", "functional", false, TestInsertAndSelectSmallRAW},
	{"TestInsertAndSelectLargeRAW", "functional", false, TestInsertAndSelectLargeRAW},
	{"TestInsertAndSelectNullRAW", "functional", false, TestInsertAndSelectNullRAW},
	{"TestRAWMultipleRows", "functional", false, TestRAWMultipleRows},
	{"TestRAWUpdateOperation", "functional", false, TestRAWUpdateOperation},
	{"TestRAWTypeSystemIntegration", "functional", false, TestRAWTypeSystemIntegration},
	{"TestDriver_Prepared_InsertMultipleRows_Re_exec", "functional", false, TestDriver_Prepared_InsertMultipleRows_Re_exec},
	{"TestDriver_Bind_ReusedNamedParameter", "functional", false, TestDriver_Bind_ReusedNamedParameter},
	{"TestDriver_Prepared_SelectMultipleRows_Re_exec_BindTypeChange", "functional", false, TestDriver_Prepared_SelectMultipleRows_Re_exec_BindTypeChange},
	{"TestDriver_Prepared_SelectMultipleRows_Re_exec", "functional", false, TestDriver_Prepared_SelectMultipleRows_Re_exec},

	{"TestCommit", "functional", false, TestCommit},
	{"TestRollback", "functional", false, TestRollback},
	{"TestRollbackThroughContextServerSleep", "functional", false, TestRollbackThroughContextServerSleep},
	{"TestRollbackThroughContextCancel", "functional", false, TestRollbackThroughContextCancel},

	{"TestDriver_Prepared_Insert_Clob_Small", "functional", false, TestDriver_Prepared_Insert_Clob_Small},
	{"TestDriver_Prepared_Insert_Clob_Large", "functional", false, TestDriver_Prepared_Insert_Clob_Large},
	{"TestDriver_Prepared_Insert_Blob_Small", "functional", false, TestDriver_Prepared_Insert_Blob_Small},
	{"TestDriver_Prepared_Insert_Blob_Large", "functional", false, TestDriver_Prepared_Insert_Blob_Large},

	{"TestReadOnlyTransaction", "functional", false, TestReadOnlyTransaction},

	{"TestDriver_Table_Select_JSON", "functional", false, TestDriver_Table_Select_JSON},
	{"TestDriver_Table_Select_NullJSON", "functional", false, TestDriver_Table_Select_NullJSON},
	{"TestDriver_Table_Select_CLOB", "functional", false, TestDriver_Table_Select_CLOB},
	{"TestDriver_Table_Select_BLOB", "functional", false, TestDriver_Table_Select_BLOB},
	{"TestDriver_Table_Insert_Select_CLOB_BLOB_MultiRows", "functional", false, TestDriver_Table_Insert_Select_CLOB_BLOB_MultiRows},
	{"TestDriver_Table_Insert_Select_JSON_MultiRows", "functional", false, TestDriver_Table_Insert_Select_JSON_MultiRows},

	{"TestDriver_DMLReturning_Insert_Ordinal", "functional", false, TestDriver_DMLReturning_Insert_Ordinal},
	{"TestDriver_DMLReturning_Update_Named", "functional", false, TestDriver_DMLReturning_Update_Named},
	{"TestDriver_DMLReturning_Insert_MultipleScalarTypes", "functional", false, TestDriver_DMLReturning_Insert_MultipleScalarTypes},
	{"TestDriver_DMLReturning_Delete_Ordinal", "functional", false, TestDriver_DMLReturning_Delete_Ordinal},
	{"TestDriver_DMLReturning_Update_Named_NoStmt", "functional", false, TestDriver_DMLReturning_Update_Named_NoStmt},
	{"TestDriver_DMLReturning_ZeroRowsAffected", "functional", false, TestDriver_DMLReturning_ZeroRowsAffected},
	{"TestDriver_DMLReturning_Delete_ZeroRowsAffected", "functional", false, TestDriver_DMLReturning_Delete_ZeroRowsAffected},
	{"TestDriver_DMLReturning_Insert_RAW", "functional", false, TestDriver_DMLReturning_Insert_RAW},
	{"TestDriver_DMLReturning_Insert_CHAR", "functional", false, TestDriver_DMLReturning_Insert_CHAR},
	{"TestDriver_DMLReturning_PreparedStmt_ReExecution", "functional", false, TestDriver_DMLReturning_PreparedStmt_ReExecution},
	{"TestDriver_DMLReturning_InTransaction_Rollback", "functional", false, TestDriver_DMLReturning_InTransaction_Rollback},
	{"TestDriver_DMLReturning_InTransaction_Commit", "functional", false, TestDriver_DMLReturning_InTransaction_Commit},
	{"TestDriver_DMLReturning_Insert_InOut", "functional", false, TestDriver_DMLReturning_Insert_InOut},
	{"TestDriver_DMLReturning_Update_MultipleRows", "functional", false, TestDriver_DMLReturning_Update_MultipleRows},
	{"TestDriver_DMLReturning_Insert_NullableColumn", "functional", false, TestDriver_DMLReturning_Insert_NullableColumn},
	{"TestDriver_DMLReturning_Insert_NullBinaryFloatIntoNullFloat64", "functional", false, TestDriver_DMLReturning_Insert_NullBinaryFloatIntoNullFloat64},
	{"TestDriver_DMLReturning_Insert_TimestampWithLocalTZ", "functional", false, TestDriver_DMLReturning_Insert_TimestampWithLocalTZ},
	{"TestDriver_DMLReturning_Insert_NumberScalePrecision", "functional", false, TestDriver_DMLReturning_Insert_NumberScalePrecision},
	{"TestDriver_DMLReturning_BinaryFloatColumn", "functional", false, TestDriver_DMLReturning_BinaryFloatColumn},
	{"TestDriver_DMLReturning_Insert_BooleanColumn", "functional", false, TestDriver_DMLReturning_Insert_BooleanColumn},
	{"TestDriver_PLSQL_InOut_NumberFunction", "functional", false, TestDriver_PLSQL_InOut_NumberFunction},
	{"TestDriver_PLSQL_InOut_VarcharProcedure", "functional", false, TestDriver_PLSQL_InOut_VarcharProcedure},
	{"TestDriver_PLSQL_ProcedureWithInOut", "functional", false, TestDriver_PLSQL_ProcedureWithInOut},
	{"TestDriver_PLSQL_ProcedureWithInOut_AllTypes", "functional", false, TestDriver_PLSQL_ProcedureWithInOut_AllTypes},
	{"TestDriver_PLSQL_InOut_NumberFunction_ReExecuteSameStatement", "functional", false, TestDriver_PLSQL_InOut_NumberFunction_ReExecuteSameStatement},
	{"TestDriver_PLSQL_InOut_NumberFunction_ReExecuteSameStatement_NamedBinds", "functional", false, TestDriver_PLSQL_InOut_NumberFunction_ReExecuteSameStatement_NamedBinds},

	{"TestDriver_PLSQL_InOut_NumberFunctionDoubleBind", "functional", false, TestDriver_PLSQL_InOut_NumberFunctionDoubleBind},

	{"TestDriver_TRIGGER_GormTest", "functional", false, TestDriver_TRIGGER_GormTest},

	{"TestDriver_SQLNullTypes_BindInputs", "functional", false, TestDriver_SQLNullTypes_BindInputs},
	{"TestDriver_SQLNullTypes_DMLReturning_OutDest", "functional", false, TestDriver_SQLNullTypes_DMLReturning_OutDest},
	{"TestDriver_SQLNullTypes_PLSQL_InOut", "functional", false, TestDriver_SQLNullTypes_PLSQL_InOut},
	{"TestDriver_SQLNullTypes_PLSQL_InOut_UsingSetupObjects", "functional", false, TestDriver_SQLNullTypes_PLSQL_InOut_UsingSetupObjects},
	{"TestDriver_SQLNullTypes_PLSQL_InOut_NullInputs", "functional", false, TestDriver_SQLNullTypes_PLSQL_InOut_NullInputs},

	{"TestConnectorConnectDisconnectsNetworkSessionWhenInstantiatorFails", "unitary", false, TestConnectorConnectDisconnectsNetworkSessionWhenInstantiatorFails},
	{"TestConnectorConnectDisconnectsNetworkSessionWhenGetConnectionFails", "unitary", false, TestConnectorConnectDisconnectsNetworkSessionWhenGetConnectionFails},
	{"TestConnectorConnectLeavesNetworkSessionOpenAfterSuccess", "unitary", false, TestConnectorConnectLeavesNetworkSessionOpenAfterSuccess},
	{"TestConnectorConnectDoesNotReturnStaleAttemptErrorAfterLaterSuccess", "unitary", false, TestConnectorConnectDoesNotReturnStaleAttemptErrorAfterLaterSuccess},
	{"TestDriver_InsertForeignKeyViolation", "functional", false, TestDriver_InsertForeignKeyViolation},
	{"TestDriver_Varchar2_TrailingSpacesPreserved", "functional", false, TestDriver_Varchar2_TrailingSpacesPreserved},
	{"TestDriver_Varchar2_EmptyStringIsNull", "functional", false, TestDriver_Varchar2_EmptyStringIsNull},
	{"TestDriver_Varchar2_EmbeddedNULRoundTrip", "functional", false, TestDriver_Varchar2_EmbeddedNULRoundTrip},
	{"TestDriver_Varchar2_BoundaryLengths", "functional", false, TestDriver_Varchar2_BoundaryLengths},
	{"TestDriver_Select_ZeroRows_FilterCondition", "functional", false, TestDriver_Select_ZeroRows_FilterCondition},
	{"TestDriver_Select_RowWithNullColumn", "functional", false, TestDriver_Select_RowWithNullColumn},
	{"TestDriver_Select_MultipleRows_SomeNulls", "functional", false, TestDriver_Select_MultipleRows_SomeNulls},
	{"TestDriver_Select_AllNullExceptPK", "functional", false, TestDriver_Select_AllNullExceptPK},
	{"TestDriver_Select_NullFromComputedExpression", "functional", false, TestDriver_Select_NullFromComputedExpression},
	{"TestDriver_OpenConnectorReturnsInvalidDSNParameterError", "unitary", false, TestDriver_OpenConnectorReturnsInvalidDSNParameterError},
	{"TestDriver_OpenConnectorStoresConnectDescriptorFromDSN", "unitary", true, TestDriver_OpenConnectorStoresConnectDescriptorFromDSN},
	{"TestDriver_OpenConnectorUsesFallbackConnectDescriptor", "unitary", false, TestDriver_OpenConnectorUsesFallbackConnectDescriptor},
	{"TestDriver_OpenConnectorUsesNSParam", "unitary", false, TestDriver_OpenConnectorUsesNSParam},
	{"TestDriver_OpenConnectorUsesNSProperty", "unitary", false, TestDriver_OpenConnectorUsesNSProperty},
	{"TestDriver_OpenConnectorUsesParam", "unitary", false, TestDriver_OpenConnectorUsesParam},
	{"TestConnection_ResetSessionKo", "functional", false, TestConnection_ResetSessionKo},
	{"TestConnection_ResetSessionOk", "functional", false, TestConnection_ResetSessionOk},
	{"TestConnection_ResetSessionPool", "functional", false, TestConnection_ResetSessionPool},
	{"TestDriver_Prepared_InsertAndSelect_AllTypes_DefaultValuesForNulls_NullScanners", "functional", false, TestDriver_Prepared_InsertAndSelect_AllTypes_DefaultValuesForNulls_NullScanners},
	{"TestDriver_Prepared_Insert_Nclob_Small", "functional", false, TestDriver_Prepared_Insert_Nclob_Small},
	{"TestDriver_Select_NumericFloatTypes_Prepared_Named", "functional", false, TestDriver_Select_NumericFloatTypes_Prepared_Named},
	{"TestDriver_TCPS_DN_Components_WhiteSpaces", "manual", false, TestDriver_TCPS_DN_Components_WhiteSpaces},
	{"TestDriver_TCPS_Handshake_EnforcesDNMatching_WhitespaceMismatchRejection", "manual", false, TestDriver_TCPS_Handshake_EnforcesDNMatching_WhitespaceMismatchRejection},
	{"TestDriver_TCPS_InvalidCertDn", "manual", false, TestDriver_TCPS_InvalidCertDn},
	{"TestDriver_TCPS_SSL_SERVER_DN_MATCH_DEFAULT", "manual", false, TestDriver_TCPS_SSL_SERVER_DN_MATCH_DEFAULT},
	{"TestDriver_TCPS_SSL_SERVER_DN_MATCH_OFF", "manual", false, TestDriver_TCPS_SSL_SERVER_DN_MATCH_OFF},
	{"TestDriver_Table_Create_Multiple_Connections", "functional", false, TestDriver_Table_Create_Multiple_Connections},
	{"TestIssue_ColumnTypeDatabaseCharTypeName", "functional", false, TestIssue_ColumnTypeDatabaseCharTypeName},
	{"TestIssue_ColumnTypeDatabaseTypeName", "functional", false, TestIssue_ColumnTypeDatabaseTypeName},
	{"TestIssue_ColumnTypePrecisionScale", "functional", false, TestIssue_ColumnTypePrecisionScale},
	{"TestIssue_DecodeBinaryColumnType", "functional", false, TestIssue_DecodeBinaryColumnType},
	{"TestServerError", "functional", false, TestServerError},
}

func TestCategoryExecutor(t *testing.T) {
	oracleTest.RunCategoryExecutor(t, oracleTest.TestCategory, testCases)
}

type Version = oracleTest.Version
type TestConfig = oracleTest.TestConfig
type TestingEnvironment = oracleTest.TestingEnvironment

var DefaultTestConfig *TestConfig
var TestEnvironement TestingEnvironment
var TestingConfig *TestConfig
