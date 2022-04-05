package main

func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	// Initialize the router
	p.router = p.InitAPI()
	return nil
}
