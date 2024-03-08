package newdata

// SQLInit performs initialization such as loading caches
func (p *DatabaseSQL) SQLInit() error {
	err := p.ParentDB.Mim.Init(p.ParentDB)
	if err != nil {
		return err
	}
	if err = p.LoadLocaleCache(); err != nil {
		return err
	}
	p.MetricIDCache = make(map[string]int, 10) // enough to get it started
	return nil
}
