package moduleindex

// GetModuleList returns a module list from storage.
//
// swagger:operation GET /registry/docs/modules/index.json Modules GetModuleList
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: A list of all modules.
//     schema:
//       '$ref': '#/definitions/ModuleList'

// GetModule returns a list of all versions of a module.
//
// swagger:operation GET /registry/docs/modules/{namespace}/{name}/{target}/index.json Modules GetModule
// ---
// produces:
// - application/json
// parameters:
// - name: namespace
//   in: path
//   required: true
//   description: Namespace of the module, all lower case.
//   type: string
// - name: name
//   in: path
//   required: true
//   description: Name of the module, all lower case.
//   type: string
// - name: target
//   in: path
//   required: true
//   description: Target system of the module, all lower case.
//   type: string
// responses:
//   '200':
//     description: A list of all versions of a module with metadata.
//     schema:
//       '$ref': '#/definitions/Module'

// GetModuleVersion returns the details of one specific module version.
//
// swagger:operation GET /registry/docs/modules/{namespace}/{name}/{target}/{version}/index.json Modules GetModuleVersion
// ---
// produces:
// - application/json
// parameters:
// - name: namespace
//   in: path
//   required: true
//   description: Namespace of the module, all lower case.
//   type: string
// - name: name
//   in: path
//   required: true
//   description: Name of the module, all lower case.
//   type: string
// - name: target
//   in: path
//   required: true
//   description: Target system of the module, all lower case.
//   type: string
// - name: version
//   in: path
//   required: true
//   description: Version number of the module with the "v" prefix.
//   type: string
// responses:
//   '200':
//     description: The details of a specific module version.
//     schema:
//       '$ref': '#/definitions/ModuleVersion'

// GetModuleReadme returns the readme of a module.
//
// swagger:operation GET /registry/docs/modules/{namespace}/{name}/{target}/{version}/README.md Modules GetModuleReadme
// ---
// parameters:
// - name: namespace
//   in: path
//   required: true
//   description: Namespace of the module, all lower case.
//   type: string
// - name: name
//   in: path
//   required: true
//   description: Name of the module, all lower case.
//   type: string
// - name: target
//   in: path
//   required: true
//   description: Target system of the module, all lower case.
//   type: string
// - name: version
//   in: path
//   required: true
//   description: Version number of the module with the "v" prefix.
//   type: string
// produces:
// - text/markdown
// responses:
//   '200':
//     description: The contents of the document.
//     schema:
//       type: file

// GetSubmoduleReadme returns the readme of a submodule.
//
// swagger:operation GET /registry/docs/modules/{namespace}/{name}/{target}/{version}/modules/{submodule}/README.md Modules GetSubmoduleReadme
// ---
// parameters:
// - name: namespace
//   in: path
//   required: true
//   description: Namespace of the module, all lower case.
//   type: string
// - name: name
//   in: path
//   required: true
//   description: Name of the module, all lower case.
//   type: string
// - name: target
//   in: path
//   required: true
//   description: Target system of the module, all lower case.
//   type: string
// - name: version
//   in: path
//   required: true
//   description: Version number of the module with the "v" prefix.
//   type: string
// - name: submodule
//   in: path
//   required: true
//   description: Submodule name.
//   type: string
// produces:
// - text/markdown
// responses:
//   '200':
//     description: The contents of the document.
//     schema:
//       type: file

// GetModuleExampleReadme returns the readme of a module example.
//
// swagger:operation GET /registry/docs/modules/{namespace}/{name}/{target}/{version}/examples/{example}/README.md Modules GetModuleExampleReadme
// ---
// parameters:
// - name: namespace
//   in: path
//   required: true
//   description: Namespace of the module, all lower case.
//   type: string
// - name: name
//   in: path
//   required: true
//   description: Name of the module, all lower case.
//   type: string
// - name: target
//   in: path
//   required: true
//   description: Target system of the module, all lower case.
//   type: string
// - name: version
//   in: path
//   required: true
//   description: Version number of the module with the "v" prefix.
//   type: string
// - name: example
//   in: path
//   required: true
//   description: Example name.
//   type: string
// produces:
// - text/markdown
// responses:
//   '200':
//     description: The contents of the document.
//     schema:
//       type: file
