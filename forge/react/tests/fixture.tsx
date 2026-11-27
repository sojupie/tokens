import { useState } from 'react';
import { createRoot } from 'react-dom/client';
import { Button, Card, Dialog, IconButton, List, ListItemButton, MaterialSymbol, NavigationItem, NavigationRail, SplitButton, TextField } from '@token-catalog/forge-react';
import '@token-catalog/forge-react/styles.css';

function Fixture() {
  const [outer, setOuter] = useState(false);
  const [inner, setInner] = useState(false);
  const [closed, setClosed] = useState(0);
  const [innerClosed, setInnerClosed] = useState(0);
  const [selected, setSelected] = useState(true);
  const [menuOpen, setMenuOpen] = useState(false);
  return <main data-component-assets>
    <Button id="open-outer" onClick={() => setOuter(true)}>Open outer</Button>
    <output id="dialog-state">{JSON.stringify({ outer, inner, closed, innerClosed })}</output>
    <Dialog id="outer-dialog" title="Outer" open={outer} onOpenChange={setOuter} onAfterClose={() => setClosed(value => value + 1)}>
      <TextField id="outer-draft" label="Draft" defaultValue="Retained input" />
      <Button id="open-inner" onClick={() => setInner(true)}>Open inner</Button>
      <Button id="close-outer" data-dialog-close="">Close outer</Button>
      <Dialog id="inner-dialog" title="Inner" open={inner} onOpenChange={setInner} onAfterClose={() => setInnerClosed(value => value + 1)}>
        <Button id="close-inner" data-dialog-close="">Close inner</Button>
        <Button id="controlled-inner" onClick={() => setInner(false)}>Close through React</Button>
      </Dialog>
    </Dialog>
    <NavigationRail label="Glyph states">
      <NavigationItem id="glyph-item" label="Files" icon={<MaterialSymbol name="layers" />} aria-current={selected ? 'page' : undefined} onClick={() => setSelected(value => !value)} />
    </NavigationRail>
    <section id="list-fixture" style={{ inlineSize: '16rem' }}>
      <List id="single-list" data-shape="baseline">
        <li><ListItemButton id="single-item" leading={<MaterialSymbol name="folder" />}>A single row</ListItemButton></li>
      </List>
      <List id="action-list" data-actions-layout="inline">
        <li><ListItemButton id="action-item" leading={<MaterialSymbol name="folder" />} actions={<IconButton id="row-action" label="Actions"><MaterialSymbol name="more_horiz" /></IconButton>}>
          A long group label that wraps in its own text column
        </ListItemButton></li>
      </List>
    </section>
    <section id="field-fixture" style={{ inlineSize: '20rem' }}>
      <TextField id="field" label="Name" leading={<MaterialSymbol name="search" />} trailing={<IconButton label="Clear"><MaterialSymbol name="close" /></IconButton>} defaultValue="Example" />
    </section>
    <div id="card-parent" style={{ color: 'var(--md-sys-color-error)' }}>
      <Card as="section" headline="Card foreground" supporting="Uses the surface role." />
    </div>
    <SplitButton label="Split states" primary={<Button id="split-primary" variant="tonal" className="md-split-leading">Action</Button>}
      menu={<Button id="split-menu" variant="tonal" className="md-split-trailing" aria-label="Menu" aria-expanded={menuOpen} onClick={() => setMenuOpen(value => !value)}><MaterialSymbol name="expand_more" /></Button>} />
  </main>;
}

const root = document.getElementById('root');
if (!root) throw new Error('Missing fixture root.');
createRoot(root).render(<Fixture />);
