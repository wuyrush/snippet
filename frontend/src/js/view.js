import React, { Component } from 'react';
import ReactDOM from 'react-dom';
import Prism from 'prismjs';

// TODO: Bloomer for React Bulma boilerplates
import {
  Content,
  Container,
  Button,
  Level, LevelLeft, LevelRight, LevelItem,
  Title,
  Notification,
  Tag,
} from 'bloomer';

const log = console,
  VOIDS = [undefined, null],
  MODE_TO_PRISM_LANG = {
    python: 'python',
    golang: 'go',
    rust: 'rust',
    javascript: 'javascript',
    text: 'textile',
};

function nonNull(obj) {
  return VOIDS.indexOf(obj) == -1;
}

function modeToPrismLang(mode) {
  return MODE_TO_PRISM_LANG[mode];
}
class ViewSnippetContainer extends Component {
  constructor(props) {
    super(props);
    this.state = {
      mode: 'python',
      snippetText: '',
      snippetName: 'Unknown',
      timeCreated: 0,
      timeExpired: 0,
      error: null,
    };
  }

  componentDidMount() {
    // fetch snippet data from backend
    fetch(''.concat('http://', document.location.host, '/api', document.location.pathname)).then(
      resp => Promise.all([Promise.resolve(resp.status), resp.text()]),
      err => {
        console.log('Failed to fetch snippet data due to network error.', err);
        // directly throw the error so that we fail the promise chain fast -- aka don't let
        // execution go to the subsequent `then` blocks.
        throw err;
      }
    ).then(
      ([responseStatus, responseBody]) => {
        if (responseStatus == 200) {
          let snippetData = JSON.parse(responseBody);
          // Need to highlight the snippet after we load it.
          // See https://reactjs.org/docs/react-component.html#setstate
          this.setState(snippetData, () => Prism.highlightAll());
        } else {
          throw `Error ${responseStatus}: ${responseBody}`;
        }
      },
      err => {
        console.log(err);
        throw err;
      }
    ).catch(err => this.setState({ error: err }));  // update UI with error we got at the end
  }

  render() {
    if (nonNull(this.state.error)) {
      return (
        <Container>
          <Notification isColor='danger'>
            { this.state.error }
          </Notification>
        </Container>
      )
    }

    return (
      <Container>
        <Title>{ this.state.snippetName }</Title>
        <Level>
          <LevelLeft>
            <LevelItem>
              <Tag isColor='info'>
                Created at { (new Date(this.state.timeCreated * 1000)).toUTCString() }
              </Tag>
            </LevelItem>
            <LevelItem>
              <Tag isColor='danger'>
                Created at { (new Date(this.state.timeExpired * 1000)).toUTCString() }
              </Tag>
            </LevelItem>
          </LevelLeft>
          <LevelRight>
            <LevelItem>
              <Button isColor="primary">Copy to clipboard</Button>
            </LevelItem>
            <LevelItem>
              <Button isColor="primary">Download</Button>
            </LevelItem>
            <LevelItem>
              <Button isColor="primary">Edit</Button>
            </LevelItem>
          </LevelRight>
        </Level>
        <Content>
          <pre>
            <code className={`language-${modeToPrismLang(this.state.mode)}`}>
              { this.state.snippetText }
            </code>
          </pre>
        </Content>
      </Container>
    )
  }
}

ReactDOM.render(<ViewSnippetContainer />, document.getElementById('root'));

