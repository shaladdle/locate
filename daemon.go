package locate

type daemon struct {
}

func NewDaemon(root string) Locator {
	return nil
}

func (d *daemon) Locate() []Result {
	return nil
}
