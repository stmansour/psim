package newdata

// SQLInit performs initialization such as loading caches
func (p *DatabaseSQL) SQLInit() error {
	err := p.ParentDB.Mim.loadMInfluencerSubclassesSQL()
	if err != nil {
		return err
	}
	if err = p.LoadLocaleCache(); err != nil {
		return err
	}
	return nil
}
