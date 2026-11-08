package gosqlbuilder

import "testing"

func BenchmarkSQLBuilder_SelectQuery(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = Select("id", "name", "email", "created_at").
			From("users").
			Where("status = ?", "ACTIVE").
			Where("role = ?", "ADMIN").
			OrderBy("created_at", "DESC").
			Limit(20).
			Offset(40).
			Build()
	}
}

func BenchmarkSQLBuilder_InsertQuery(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = InsertInto("orders").
			Columns("user_id", "total_amount", "currency", "status").
			Values("usr_123", 45000, "INR", "PENDING").
			Build()
	}
}
