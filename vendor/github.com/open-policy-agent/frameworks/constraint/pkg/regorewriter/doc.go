/*
package regorewriter rewrites import and package refs for a set of rego modules.

Rego modules are divided into two categories: libraries and constraint templates.  The libraries
will have both package path, imports and data references updated while the constraint templates
will only have imports and data references updated.
*/
package regorewriter
