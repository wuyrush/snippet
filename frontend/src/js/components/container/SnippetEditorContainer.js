/* Top-level component for snippet editor page. */
import React, { Component } from 'react';
import ReactDOM from 'react-dom';
import brace from 'brace';
import AceEditor from 'react-ace';

class SnippetEditorContainer extends Component {
  constructor(props) {
    super(props);
    this.state = {
    }
  }

  render() {
    return (
      <>
        <SnippetNameInput value='FakeInput' />
        <div class="columns">
          <div class="column is-2">
            <SelectDropdown options={['python', 'golang', 'rust']} label='Syntax' />
          </div>
          <div class="column is-2">
            <SelectDropdown options={['terminal', 'monokai', 'github']} label='Theme' />
          </div>
        </div>
        <SnippetEditorInput />
        <div class="field is-grouped">
          <div class="control">
            <Button className="button is-link" label="Save" />
          </div>
          <div class="control">
            <Button className="button is-primary" label="Lock" />
          </div>
          <div class="control">
            <Button className="button is-primary" label="Copy to Clipboard" />
          </div>
          <div class="control">
            <Button className="button is-primary" label="Download" />
          </div>
        </div>
      </>
    )
  }
}

function SnippetNameInput(props) {
  return (
    <div class="field">
      <label class="label">Snippet Name</label>
      <div class="control">
        <input class="input is-primary" type="text" value={props.value} placeholder="Snippet name" />
      </div>
    </div>
  )
}

function SelectDropdown(props) {
  return (
    <div class="field is-horizontal">
      <div class="field-label is-normal">
        <label class="label">{props.label}</label>
      </div>
      <div class="field-body">
        <div class="field">
          <div class="control">
            <div class="select is-primary">
              <select>
                {
                  props.options.map(opt => (
                    <option key={opt} value={opt}>
                      {opt}
                    </option>
                  ))
                }
              </select>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function Checkbox(props) {
  return (
    <div class="field">
      <div class="control">
        <label class="checkbox">
          <input type="checkbox" />
          {props.label}
        </label>
      </div>
    </div>
  )
}

function SnippetEditorInput(props) {
  return (
    <div class="container">
      <AceEditor
        tabSize={2}
        highlightActiveLine={true}
        height='500px'
        width='100%'
        fontSize={14}
      />
    </div>
  )
}

function Button(props) {
  return <a className={props.className}>{props.label}</a>
}

ReactDOM.render(<SnippetEditorContainer />, document.getElementById('root'));

export default SnippetEditorContainer;
