package filter

import rego.v1

default match := false

match if {
	not listAvailableExtension
	not fuzzySearch
	not installStatusSearch
	not enabledStatusSearch
}

match if {
	listAvailableExtension
}

match if {
	listAvailableExtension
	alreadyInstalled
}

match if {
	listAvailableExtension
}

match if {
	fuzzySearch
	displayNameMatch
}

match if {
	installStatusSearch
	installStatusMatch
}

match if {
	installStatusSearch
	installedStatusMatch
}

match if {
	enabledStatusSearch
	enabledStatusMatch
}

match if {
	installStatusSearch
	notInstalledStatusMatch
}

fuzzySearch if "q" == input.filter.field

installStatusSearch if "status" == input.filter.field

enabledStatusSearch if "enabled" == input.filter.field

listAvailableExtension if "available" == input.filter.field

alreadyInstalled if input.object.status.state != ""

displayNameMatch if {
	contains(lower(input.object.spec.displayName[_]), lower(input.filter.value))
}

nameMatch if {
	contains(lower(input.object.metadata.name), lower(input.filter.value))
}

installStatusMatch if {
	lower(input.object.status.state) == lower(input.filter.value)
}

installedStatusMatch if {
	input.filter.value == "installed"
	"Installed" == input.object.status.state
}

enabledStatusMatch if {
	alreadyInstalled
	input.filter.value == "true"
	input.object.status.enabled
}

enabledStatusMatch if {
	alreadyInstalled
	input.filter.value == "false"
	not input.object.status.enabled
}

notInstalledStatusMatch if {
    input.filter.value == "notInstalled"
    not alreadyInstalled
}
