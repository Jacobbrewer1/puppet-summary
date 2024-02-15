package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/smallfish/simpleyaml"
)

// parseHost reads the `host` parameter from the YAML and populates
// the given report-structure with suitable values.
func parseHost(y *simpleyaml.Yaml, out *entities.PuppetReport) error {
	host, err := y.Get("host").String()
	if err != nil {
		return errors.New("failed to get 'host' from YAML")
	}
	reg, _ := regexp.Compile("^([a-z0-9._-]+)$")
	if !reg.MatchString(host) {
		return errors.New("the submitted 'host' field failed our security check")
	}
	out.Fqdn = host
	return nil
}

// parseEnvironment reads the `environment` parameter from the YAML and populates
// the given report-structure with suitable values.
func parseEnvironment(y *simpleyaml.Yaml, out *entities.PuppetReport) error {
	envStr, err := y.Get("environment").String()
	if err != nil {
		return errors.New("failed to get 'environment' from YAML")
	}
	reg, _ := regexp.Compile("^([A-Za-z0-9_]+)$")
	if !reg.MatchString(envStr) {
		return errors.New("the submitted 'environment' field failed our security check")
	}
	env := entities.Environment(strings.ToUpper(envStr))
	if !env.Valid() {
		return fmt.Errorf("invalid environment '%s'", env)
	}
	out.Env = env
	return nil
}

// parseTime reads the `time` parameter from the YAML and populates
// the given report-structure with suitable values.
func parseTime(y *simpleyaml.Yaml, out *entities.PuppetReport) error {
	// Get the time puppet executed
	at, err := y.Get("time").String()
	if err != nil {
		return errors.New("failed to get 'time' from YAML")
	}

	// Strip any quotes that might surround the time.
	at = strings.Replace(at, "'", "", -1)

	// Convert "T" -> " "
	at = strings.Replace(at, "T", " ", -1)

	// strip the time at the first period.
	parts := strings.Split(at, ".")
	at = parts[0]

	// Parse 2017-07-29 23:17:01 as a time.
	t, err := time.Parse("2006-01-02 15:04:05", at)
	if err != nil {
		return fmt.Errorf("failed to parse time '%s' as time", at)
	}

	out.ExecTime = entities.Datetime(t)

	return nil
}

// parseStatus reads the `status` parameter from the YAML and populates
// the given report-structure with suitable values.
func parseStatus(y *simpleyaml.Yaml, out *entities.PuppetReport) error {
	s, err := y.Get("status").String()
	if err != nil {
		return errors.New("failed to get 'status' from YAML")
	}

	state := entities.State(strings.ToUpper(s))
	if !state.Valid() {
		return fmt.Errorf("invalid state '%s'", state)
	}

	out.State = state
	return nil
}

// parseRuntime reads the `metrics.time.values` parameters from the YAML
// and populates given report-structure with suitable values.
func parseRuntime(y *simpleyaml.Yaml, out *entities.PuppetReport) error {
	times, err := y.Get("metrics").Get("time").Get("values").Array()
	if err != nil {
		return err
	}

	r, _ := regexp.Compile("Total ([0-9.]+)")

	runtime := ""
	for _, value := range times {
		match := r.FindStringSubmatch(fmt.Sprint(value))
		if len(match) == 2 {
			runtime = match[1]
		}
	}

	// Parse the runtime as a duration.
	d, err := time.ParseDuration(runtime + "s")
	if err != nil {
		return fmt.Errorf("failed to parse runtime '%s' as duration", runtime)
	}

	out.Runtime = entities.Duration(d)

	return nil
}

// parseResources looks for the counts of resources which have been
// failed, changed, skipped, etc, and updates the given report-structure
// with those values.
func parseResources(y *simpleyaml.Yaml, out *entities.PuppetReport) error {
	resources, err := y.Get("metrics").Get("resources").Get("values").Array()
	if err != nil {
		return fmt.Errorf("failed to get 'metrics.resources.values' from YAML: %w", err)
	}

	totalReg, err := regexp.Compile("Total ([0-9.]+)")
	if err != nil {
		return fmt.Errorf("failed to compile total regexp: %w", err)
	}
	failedReg, _ := regexp.Compile("Failed ([0-9.]+)")
	if err != nil {
		return fmt.Errorf("failed to compile failed regexp: %w", err)
	}
	skippedReg, err := regexp.Compile("Skipped ([0-9.]+)")
	if err != nil {
		return fmt.Errorf("failed to compile skipped regexp: %w", err)
	}
	changedReg, _ := regexp.Compile("Changed ([0-9.]+)")
	if err != nil {
		return fmt.Errorf("failed to compile changed regexp: %w", err)
	}

	total := ""
	changed := ""
	failed := ""
	skipped := ""

	for _, value := range resources {
		mr := totalReg.FindStringSubmatch(fmt.Sprint(value))
		if len(mr) == 2 {
			total = mr[1]
		}
		mf := failedReg.FindStringSubmatch(fmt.Sprint(value))
		if len(mf) == 2 {
			failed = mf[1]
		}
		ms := skippedReg.FindStringSubmatch(fmt.Sprint(value))
		if len(ms) == 2 {
			skipped = ms[1]
		}
		mc := changedReg.FindStringSubmatch(fmt.Sprint(value))
		if len(mc) == 2 {
			changed = mc[1]
		}
	}

	// Convert the strings to integers.
	totalInt, err := strconv.Atoi(total)
	if err != nil {
		return fmt.Errorf("failed to convert total '%s' to integer", total)
	}
	out.Total = int64(totalInt)
	changedInt, err := strconv.Atoi(changed)
	if err != nil {
		return fmt.Errorf("failed to convert changed '%s' to integer", changed)
	}
	out.Changed = int64(changedInt)
	failedInt, err := strconv.Atoi(failed)
	if err != nil {
		return fmt.Errorf("failed to convert failed '%s' to integer", failed)
	}
	out.Failed = int64(failedInt)
	skippedInt, err := strconv.Atoi(skipped)
	if err != nil {
		return fmt.Errorf("failed to convert skipped '%s' to integer", skipped)
	}
	out.Skipped = int64(skippedInt)
	return nil
}

// parseLogs updates the given report with any logged messages.
func parseLogs(y *simpleyaml.Yaml, out *entities.PuppetReport) error {
	logs, err := y.Get("logs").Array()
	if err != nil {
		return errors.New("failed to get 'logs' from YAML")
	}

	logged := make([]string, 0)

	for _, v2 := range logs {
		// create a map
		m := make(map[string]string)
		v := reflect.ValueOf(v2)
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				strct := v.MapIndex(key)

				// Store the key/val in the map.
				key, val := key.Interface(), strct.Interface()
				m[key.(string)] = fmt.Sprint(val)
			}
		}
		if len(m["message"]) > 0 {
			logged = append(logged, m["source"]+" : "+m["message"])
		}
	}

	out.LogMessages = logged
	return nil
}

// parseResults updates the given report with details of any resource
// which was failed, changed, or skipped.
func parseResults(y *simpleyaml.Yaml, out *entities.PuppetReport) error {
	rs, err := y.Get("resource_statuses").Map()
	if err != nil {
		return errors.New("failed to get 'resource_statuses' from YAML")
	}

	failed := make([]*entities.PuppetResource, 0)
	changed := make([]*entities.PuppetResource, 0)
	skipped := make([]*entities.PuppetResource, 0)
	ok := make([]*entities.PuppetResource, 0)

	for _, v2 := range rs {
		m := make(map[string]string)
		v := reflect.ValueOf(v2)
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				strct := v.MapIndex(key)

				// Store the key/val in the map.
				k, v := key.Interface(), strct.Interface()
				m[k.(string)] = fmt.Sprint(v)
			}
		}

		// Now we should be able to look for skipped ones.
		if m["skipped"] == "true" {
			skipped = append(skipped,
				&entities.PuppetResource{Name: m["title"],
					Type: m["resource_type"],
					File: m["file"],
					Line: m["line"],
				},
			)
		}

		// Now we should be able to look for skipped ones.
		if m["changed"] == "true" {
			changed = append(changed,
				&entities.PuppetResource{Name: m["title"],
					Type: m["resource_type"],
					File: m["file"],
					Line: m["line"],
				},
			)
		}

		// Now we should be able to look for skipped ones.
		if m["failed"] == "true" {
			failed = append(failed,
				&entities.PuppetResource{
					Name: m["title"],
					Type: m["resource_type"],
					File: m["file"],
					Line: m["line"],
				},
			)
		}

		if m["failed"] == "false" &&
			m["skipped"] == "false" &&
			m["changed"] == "false" {
			ok = append(ok,
				&entities.PuppetResource{Name: m["title"],
					Type: m["resource_type"],
					File: m["file"],
					Line: m["line"]})
		}

	}

	out.ResourcesSkipped = skipped
	out.ResourcesFailed = failed
	out.ResourcesChanged = changed
	out.ResourcesOK = ok

	return nil
}

func parsePuppetVersion(y *simpleyaml.Yaml, out *entities.PuppetReport) error {
	version, err := y.Get("puppet_version").String()
	if err != nil {
		return errors.New("failed to get 'puppet_version' from YAML")
	}

	// Strip any quotes that might surround the version.
	version = strings.Replace(version, "'", "", -1)

	// Trim the version to 1 decimal place. (4.8.2 -> 4.8)
	elms := strings.Split(version, ".")
	if len(elms) > 2 {
		version = elms[0] + "." + elms[1]
	}

	// Convert the version to a float.
	v, err := strconv.ParseFloat(version, 64)
	if err != nil {
		return fmt.Errorf("failed to parse puppet_version '%s' as float", version)
	}

	out.PuppetVersion = v

	return nil
}

// parsePuppetReport is our main function in this module. Given an
// array of bytes we read the input and produce a PuppetReport structure.
func parsePuppetReport(content []byte) (*entities.PuppetReport, error) {
	rep := new(entities.PuppetReport)

	yaml, err := simpleyaml.NewYaml(content)
	if err != nil {
		return rep, errors.New("failed to parse YAML")
	}

	err = parseHost(yaml, rep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host: %w", err)
	}

	err = parsePuppetVersion(yaml, rep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse puppet_version: %w", err)
	}

	versionStr := fmt.Sprintf("%f", rep.PuppetVersion)
	if versionStr == "0.0" {
		return nil, fmt.Errorf("failed to parse puppet_version: %w", err)
	}

	err = parseEnvironment(yaml, rep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment: %w", err)
	}

	err = parseTime(yaml, rep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time: %w", err)
	}

	err = parseStatus(yaml, rep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	err = parseRuntime(yaml, rep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse runtime: %w", err)
	}

	err = parseResources(yaml, rep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resources: %w", err)
	}

	err = parseLogs(yaml, rep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse logs: %w", err)
	}

	err = parseResults(yaml, rep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse results: %w", err)
	}

	rep.SortResources()

	// Store the SHA1-hash of the content as the ID. This is used to detect duplicate submissions.
	helper := sha1.New()
	helper.Write(content)
	rep.ID = fmt.Sprintf("%x", helper.Sum(nil))

	return rep, nil
}
