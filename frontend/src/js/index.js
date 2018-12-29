/* Top-level component for snippet editor page. */
import React, { Component } from 'react';
import ReactDOM from 'react-dom';
import brace from 'brace';
import AceEditor from 'react-ace';

// Bloomer for React Bulma boilerplates
import {
  Columns,
  Column,
  Field,
  Control,
  Label,
  Select,
  Input,
  Container,
  FieldLabel,
  FieldBody,
  Checkbox,
  Button,
  Notification,
} from 'bloomer';

const log = console,
// TODO: at some point we need to figure out how to dynamically import these modes and themes
// otherwise the editor page will be loaded with too much unnecessary stuff(though the user might
// need them later ... We can have browsers cache them so that only the first load is slow - but
// keep in mind that slow first load can already kill many user's interest:(  )
  modes = ['python', 'golang', 'rust', 'javascript', 'text'],
  themes = ['terminal'];

modes.forEach(mode => {
  require(`brace/mode/${mode}`);
});

themes.forEach(theme => {
  require(`brace/theme/${theme}`);
});

function nonNull(obj) {
  return [undefined, null].indexOf(obj) == -1;
}

class SnippetEditorContainer extends Component {
  constructor(props) {
    super(props);
    this.state = {
      snippetName: '',
      snippetText: '',
      mode: 'python',
      theme: 'terminal',
      editorLocked: false,
      notification: null, // bloom notification. ex: { color: 'success', message: 'snippet saved.' }
      savedSnippets: [],  // list of ids of saved snippets
    };

    // bind the component instance itself with the function - so in the future when we pass
    // `this.handleChange` as function object(aka `let someFunc = this.handleChange; someFunc(...)`)
    // it will always be called with thisValue set to the component instance. This is why we can
    // pass the functions defined for the parent component down to its children and enable the
    // parent component to update states whenever children component changes.
    this.handleChange = this.handleChange.bind(this);
    this.handleEditorTextChange = this.handleEditorTextChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  // general purpose handler for inputs like text / select / Checkbox etc. The idea is container
  // component must have *full* control of the input / data used to render presentational components
  // associated with it - essentially the ids of widgets which can contain mutable states and their
  // values. Only by this can the container component know how to map(aka update its states).
  handleChange(event) {
    log.info(`handleChange: Got event. [id=${event.target.id}, val=${event.target.value}]`);
    this.setState({
      [event.target.id]: event.target.value
    })
  }

  handleEditorTextChange(text, event) {
    this.setState({snippetText: text});
  }

  handleSubmit(event) {
    // take control on browser behavior on such event by dropping the default behavior
    event.preventDefault();
    // use React as integration point since we let it manage all the UI states
    let fd = new FormData();
    ['snippetName', 'snippetText', 'mode'].forEach(name => fd.append(name, this.state[name]));

    // fire a POST request with fetch
    let postUrl = ''.concat('http://', document.location.host, '/api/save');
    fetch(postUrl, { method: 'POST', body: fd }).then(
      resp => Promise.all([Promise.resolve(resp.status), resp.text()]),
      err => {
        // network or CORS error
        log.error('Got unexpected error when submitting snippet data:', err);
        throw 'Failed to save snippet data: ' + err;
      }
    ).then(
      ([responseStatus, responseBody]) => {
        log.info('Response status:', responseStatus, 'Response body:', responseBody);
        if (responseStatus == 200) {
          let savedSnippetId = JSON.parse(responseBody)['snippetId'];
          this.setState((state, props) => {
            return {
              notification: { color: 'primary', message: 'snippet saved.' },
              savedSnippets: state.savedSnippets.concat([savedSnippetId]) 
            }
          });
          return
        }
        throw `Error ${responseStatus}: ${responseBody}`;
      },
      err => {
        log.error(err);
        throw err;
      }
    ).catch(err => {
      // notify user
      this.setState({
        notification: { color: 'danger', message: err  }
      })
    });
  }

  render() {
    let savedSnippetsJSX = null,
      savedSnippets = this.state.savedSnippets;
    if (savedSnippets.length != 0) {
      savedSnippetsJSX = (
        <Container>
        <br />
        {
          [...savedSnippets].reverse().map(sid => (
          <Notification isColor='primary'>
            <a href={"/view/" + sid}>Click to view saved snippet {sid}</a>
          </Notification>
          ))
        }
        </Container>
      )
    }

    return (
      <Container>
        {nonNull(this.state.notification) &&
          <Notification isColor={this.state.notification.color}>
            {this.state.notification.message}
          </Notification>
        }
        <form onSubmit={this.handleSubmit}>
          <InputField
            id='snippetName' handleChange={this.handleChange}
            label='Snippet Name' placeholder='dont-say-i-dont-know'
          />
          <Columns isGrid>
            <Column isSize="narrow">
              <SelectField
                options={modes}
                label='Mode' id='mode' value={this.state.mode}
                handleChange={this.handleChange}
              />
            </Column>
            <Column isSize="narrow">
              <SelectField
                options={themes}
                label='Theme' id='theme' value={this.state.theme}
                handleChange={this.handleChange}
              />
            </Column>
          </Columns>
          <SnippetEditorInput
            mode={this.state.mode}
            theme={this.state.theme}
            text={this.state.snippetText}
            handleChange={this.handleEditorTextChange}
          />
          <br/>
          <Columns isGrid>
            <Column isSize="narrow">
              <Button isColor="link" type="submit">Save</Button>
            </Column>
            <Column isSize="narrow">
              <Button isColor="primary">Lock</Button>
            </Column>
            <Column isSize="narrow">
              <Button isColor="primary">Clear</Button>
            </Column>
            <Column isSize="narrow">
              <Button isColor="primary">Copy to clipboard</Button>
            </Column>
            <Column isSize="narrow">
              <Button isColor="primary">Download</Button>
            </Column>
          </Columns>
        </form>
        {this.state.savedSnippets.length != 0 && savedSnippetsJSX}
      </Container>
    )
  }
}

function InputField(props) {
  return (
    <Field>
      <Label>{props.label}</Label>
      <Control>
        <Input isColor="primary"
          id={props.id} onChange={props.handleChange}
          placeholder={props.placeholder} defaultValue={props.defaultValue} />
      </Control>
    </Field>
  )
}

function SelectField(props) {
  return (
    <Field isHorizontal>
      <FieldLabel isNormal>
        <Label>{props.label}</Label>
      </FieldLabel>
      <FieldBody>
        <Field>
          <Control>
            <Select isColor='primary' id={props.id} onChange={props.handleChange}>
              {
                props.options.map(opt => (
                  <option key={opt} value={opt}>{opt}</option>
                ))
              }
            </Select>
          </Control>
        </Field>
      </FieldBody>
    </Field>
  )
}

function CheckboxField(props) {
  return (
    <Field>
      <Control>
        <Checkbox>{props.label}</Checkbox>
      </Control>
    </Field>
  )
}

function SnippetEditorInput(props) {
  return (
    <Container>
      <AceEditor
        mode={props.mode}
        theme={props.theme}
        value={props.text}
        onChange={props.handleChange}
        tabSize={2}
        highlightActiveLine={true}
        height='500px'
        width='100%'
        fontSize={14}
        editorProps={{$blockScrolling: true}}
      />
    </Container>
  )
}

ReactDOM.render(<SnippetEditorContainer />, document.getElementById('root'));

