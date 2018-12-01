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
} from 'bloomer';

// TODO: at some point we need to figure out how to dynamically import these modes and themes
// otherwise the editor page will be loaded with too much unnecessary stuff(though the user might
// need them later ... We can have browsers cache them so that only the first load is slow - but
// keep in mind that slow first load can already kill many user's interest:(  )

const languages = [
  'python',
  'golang',
  'rust',
  'javascript',
];

languages.forEach(lang => {
  require(`brace/mode/${lang}`);
});

const themes = [
  'terminal',
];

themes.forEach(theme => {
  require(`brace/theme/${theme}`);
});

class SnippetEditorContainer extends Component {
  constructor(props) {
    super(props);
    this.state = {
      snippet_name: '',
      snippet_text: '',
      lang: 'python',
      theme: 'terminal',
      editorLocked: false,
    };

    // bind the component instance itself with the function - so in the future when we pass
    // `this.handleChange` as function object(aka `let someFunc = this.handleChange; someFunc(...)`)
    // it will always be called with thisValue set to the component instance. This is why we can
    // pass the functions defined for the parent component down to its children and enable the
    // parent component to update states whenever children component changes.
    this.handleChange = this.handleChange.bind(this);
    this.handleEditorTextChange = this.handleEditorTextChange.bind(this);
  }

  // general purpose handler for inputs like text / select / Checkbox etc. The idea is container
  // component must have *full* control of the input / data used to render presentational components
  // associated with it - essentially the ids of widgets which can contain mutable states and their
  // values. Only by this can the container component know how to map(aka update its states).
  handleChange(event) {
    console.log(`handleChange: Got event. [id=${event.target.id}, val=${event.target.value}]`);
    this.setState({
      [event.target.id]: event.target.value
    })
  }

  handleEditorTextChange(text, event) {
    this.setState({snippet_text: text});
  }
  render() {
    return (
      <>
        <InputField label='Snippet Name' placeholder='dont-say-i-dont-know' />
        <Columns isGrid>
          <Column isSize="narrow">
            <SelectField
              options={languages}
              label='Language' id='lang' value={this.state.lang}
              handleChange={this.handleChange}
            />
          </Column>
          <Column isSize="narrow">
            <SelectField
              options={themes}
              label='Theme' id='theme' value={this.state.lang}
              handleChange={this.handleChange}
            />
          </Column>
        </Columns>
        <SnippetEditorInput
          lang={this.state.lang}
          theme={this.state.theme}
          text={this.state.snippet_text}
          handleChange={this.handleEditorTextChange}
        />
        <br/>
        <Columns isGrid>
          <Column isSize="narrow">
            <Button isColor="link">Save</Button>
          </Column>
          <Column isSize="narrow">
            <Button isColor="primary">Lock</Button>
          </Column>
          <Column isSize="narrow">
            <Button isColor="primary">Copy to clipboard</Button>
          </Column>
          <Column isSize="narrow">
            <Button isColor="primary">Download</Button>
          </Column>
        </Columns>
      </>
    )
  }
}

function InputField(props) {
  return (
    <Field>
      <Label>{props.label}</Label>
      <Control>
        <Input isColor="primary" placeholder={props.placeholder} defaultValue={props.defaultValue} />
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
        mode={props.lang}
        theme={props.theme}
        value={props.text}
        onChange={props.handleChange}
        tabSize={2}
        highlightActiveLine={true}
        height='500px'
        width='100%'
        fontSize={14}
      />
    </Container>
  )
}

ReactDOM.render(<SnippetEditorContainer />, document.getElementById('root'));

export default SnippetEditorContainer;
