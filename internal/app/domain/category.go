package domain

import "fmt"

type Category struct {
	id   int
	name string
}

func NewCategory(id int, name string) (Category, error) {
	category := Category{}
	if err := category.SetId(id); err != nil {
		return category, err
	}
	if err := category.SetName(name); err != nil {
		return category, err
	}
	return category, nil
}

// Getter methods

func (c *Category) Id() int {
	return c.id
}

func (c *Category) Name() string {
	return c.name
}

// Setter methods

func (c *Category) SetId(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid category id: %d", id)
	}
	c.id = id
	return nil
}

func (c *Category) SetName(name string) error {
	if name == "" {
		return fmt.Errorf("invalid category name: %s", name)
	}
	c.name = name
	return nil
}
