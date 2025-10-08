package storage

import "testing"

func TestCondition(t *testing.T) {
	t.Run("Test toString", func(t *testing.T) {
		cond := &Condition{"name", "=", "eric"}

		should := "name = 'eric'"
		str := cond.ToString()

		if str != should {
			t.Fatalf("Expected %s get %s", should, str)
		}
	})

	t.Run("Test no field error", func(t *testing.T) {
		cond := &Condition{"name", "=", "eric"}

		entity := NewEntity()
		entity["id"] = "1"

		_, err := cond.Evaluate(entity)

		if err == nil {
			t.Fatal("Expected error with no found field")
		}
	})

	t.Run("Test unsupported operator", func(t *testing.T) {
		cond := &Condition{"name", "/", "eric"}

		entity := NewEntity()
		entity["name"] = "eric"

		_, err := cond.Evaluate(entity)
		if err == nil {
			t.Fatal("Expected error with unsupported operator")
		}
	})

	t.Run("Test = (true)", func(t *testing.T) {
		cond := &Condition{"name", "=", "eric"}

		entity := NewEntity()
		entity["name"] = "eric"

		flag, err := cond.Evaluate(entity)
		if err != nil {
			t.Fatalf("Unexpected error %s", err.Error())
		}

		if flag != true {
			t.Fatalf("Wrong answer")
		}
	})

	t.Run("Test = (false)", func(t *testing.T) {
		cond := &Condition{"name", "=", "eric"}

		entity := NewEntity()
		entity["name"] = "eri"

		flag, err := cond.Evaluate(entity)
		if err != nil {
			t.Fatalf("Unexpected error %s", err.Error())
		}

		if flag == true {
			t.Fatalf("Wrong answer")
		}
	})

	t.Run("Test LIKE (true)", func(t *testing.T) {
		cases := make([]Condition, 0)
		cases = append(cases, Condition{"name", "LIKE", "eri%"})
		cases = append(cases, Condition{"name", "LIKE", "%ri%"})
		cases = append(cases, Condition{"name", "LIKE", "%ric"})

		entity := NewEntity()
		entity["name"] = "eric"

		for _, c := range cases {
			flag, err := c.Evaluate(entity)
			if err != nil {
				t.Fatalf("Unexpected error %s", err.Error())
			}

			if flag != true {
				t.Fatalf("Wrong answer at %s", c.ToString())
			}
		}
	})

	t.Run("Test LIKE (false)", func(t *testing.T) {
		cond := &Condition{"name", "LIKE", "eri%"}

		entity := NewEntity()
		entity["name"] = "ric"

		flag, err := cond.Evaluate(entity)
		if err != nil {
			t.Fatalf("Unexpected error %s", err.Error())
		}

		if flag != true {
			t.Fatalf("Wrong answer")
		}
	})

	t.Run("Test >", func(t *testing.T) {
		cond := &Condition{"age", ">", "15"}

		entity := NewEntity()
		entity["age"] = "16"

		flag, err := cond.Evaluate(entity)
		if err != nil {
			t.Fatalf("Unexpected error %s", err.Error())
		}

		if flag != true {
			t.Fatalf("Wrong answer")
		}
	})

	t.Run("Test <", func(t *testing.T) {
		cond := &Condition{"age", "<", "15"}

		entity := NewEntity()
		entity["age"] = "14"

		flag, err := cond.Evaluate(entity)
		if err != nil {
			t.Fatalf("Unexpected error %s", err.Error())
		}

		if flag != true {
			t.Fatalf("Wrong answer")
		}
	})

	t.Run("Test IN", func(t *testing.T) {
		collection := make([]string, 0)
		collection = append(collection, "eric")
		collection = append(collection, "sam")

		cond := &Condition{"name", "IN", collection}

		entity := NewEntity()
		entity["name"] = "eric"

		flag, err := cond.Evaluate(entity)
		if err != nil {
			t.Fatalf("Unexpected error %s", err.Error())
		}

		if flag != true {
			t.Fatalf("Wrong answer")
		}
	})
}

func TestBinary(t *testing.T) {
	t.Run("Test true AND true", func(t *testing.T) {
		cond1 := &Condition{"name", "=", "eric"}
		cond2 := &Condition{"age", "=", "17"}

		entity := NewEntity()
		entity["id"] = "1"
		entity["name"] = "eric"
		entity["age"] = "17"

		query := NewQuery(cond1).And(cond2)

		flag, err := query.Evaluate(entity)
		if err != nil {
			t.Fatalf("Unexpected error %s", err.Error())
		}

		if flag != true {
			t.Fatalf("Wrong flag")
		}
	})
}
