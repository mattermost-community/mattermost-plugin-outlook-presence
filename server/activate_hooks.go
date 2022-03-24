package main

func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	// Initialize DB service
	p.router = p.InitAPI()
	return nil
}
