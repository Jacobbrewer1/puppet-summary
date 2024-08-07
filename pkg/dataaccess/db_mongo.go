package dataaccess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

const mongoDatabase = "puppet-summary"

type mongodbImpl struct {
	// client is the database.
	client *mongo.Client
}

func (m *mongodbImpl) Reconnect(ctx context.Context, connStr string) error {
	if m.client != nil {
		if err := m.client.Disconnect(ctx); err != nil {
			return fmt.Errorf("error disconnecting from database: %w", err)
		}
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(connStr).SetServerAPIOptions(serverAPI)
	opts.SetAppName(mongoDatabase)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return fmt.Errorf("connect to MongoDB: %w", err)
	} else if client == nil {
		return errors.New("nil MongoDB client")
	}

	m.client = client

	return nil
}

func (m *mongodbImpl) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func (m *mongodbImpl) Purge(ctx context.Context, from time.Time) (int, error) {
	collection := m.client.Database(mongoDatabase).Collection("reports")

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("purge"))
	defer t.ObserveDuration()

	// Delete the reports from the database out of the given range.
	res, err := collection.DeleteMany(ctx, bson.M{
		"exec_time": bson.M{
			"$lt": from.String(),
		},
	})
	if err != nil {
		return 0, fmt.Errorf("error purging data: %w", err)
	}

	return int(res.DeletedCount), nil
}

func (m *mongodbImpl) GetEnvironments(ctx context.Context) ([]summary.Environment, error) {
	collection := m.client.Database(mongoDatabase).Collection("reports")

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_environments"))

	cursor, err := collection.Distinct(ctx, "env", bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error getting environments: %w", err)
	}

	t.ObserveDuration()

	environments := make([]summary.Environment, 0)
	for _, env := range cursor {
		// Convert the cursor to a string.
		envString := fmt.Sprintf("%s", env)

		// Convert the string to an environment.
		environment := summary.Environment(envString)

		// Add the environment to the slice.
		environments = append(environments, environment)
	}

	return environments, nil
}

func (m *mongodbImpl) GetHistory(ctx context.Context, environment ...summary.Environment) ([]*entities.PuppetHistory, error) {
	// First get the distinct dates from the database.
	collection := m.client.Database(mongoDatabase).Collection("reports")

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_history"))

	envFilter := bson.M{}
	if environment != nil {
		envFilter["env"] = bson.M{
			"$in": environment,
		}
	}

	cursor, err := collection.Distinct(ctx, "exec_time", envFilter)
	if err != nil {
		return nil, fmt.Errorf("error getting history: %w", err)
	}

	t.ObserveDuration()

	datesMap := make(map[string]bool)
	for _, date := range cursor {
		// Convert the date-times to a time value.
		dateTime, err := time.Parse(time.RFC3339, date.(string))
		if err != nil {
			slog.Error("Error parsing date", slog.String(logging.KeyError, err.Error()))
			continue
		}

		// Convert the time value to a date string.
		dateString := dateTime.Format(time.DateOnly)

		// Add the date to the map.
		datesMap[dateString] = true
	}

	// Convert the map to a slice.
	dates := make([]string, 0, len(datesMap))
	for date := range datesMap {
		dates = append(dates, date)
	}

	limit := 30
	if len(dates) < limit {
		limit = len(dates)
	}

	// Sort the dates in reverse order.
	sort.Slice(dates, func(i, j int) bool {
		// Parse the dates.
		iDate, err := time.Parse(time.DateOnly, dates[i])
		if err != nil {
			slog.Error("Error parsing date", slog.String(logging.KeyError, err.Error()))
			return false
		}
		jDate, err := time.Parse(time.DateOnly, dates[j])
		if err != nil {
			slog.Error("Error parsing date", slog.String(logging.KeyError, err.Error()))
			return false
		}

		return iDate.After(jDate)
	})

	// Get the last "limit" dates.
	dates = dates[:limit]

	// Get the reports for each date.
	historyMap := make(map[string]*entities.PuppetHistory)
	for _, date := range dates {
		// Get the reports for the date. This has to be done between midnight and midnight.
		startTime, err := time.Parse(time.DateOnly, date)
		if err != nil {
			slog.Error("Error parsing date", slog.String(logging.KeyError, err.Error()))
			continue
		}

		endTime := startTime.AddDate(0, 0, 1)

		filter := envFilter
		filter["exec_time"] = bson.M{
			"$gte": startTime.Format(time.RFC3339),
			"$lt":  endTime.Format(time.RFC3339),
		}

		// For each state, count the number of reports.
		for _, state := range []summary.State{summary.State_CHANGED, summary.State_FAILED, summary.State_SKIPPED, summary.State_UNCHANGED} {
			f := filter
			f["state"] = state

			cur, err := collection.CountDocuments(ctx, f)
			if err != nil {
				return nil, fmt.Errorf("error getting history: %w", err)
			}

			// Retrieve the count for the state.
			count := int(cur)

			// Create the history object for this date if it doesn't exist.
			if _, ok := historyMap[date]; !ok {
				historyMap[date] = &entities.PuppetHistory{
					Date: date,
				}
			}

			// Add the count to the history object.
			historyMap[date].AddCount(state, count)
		}
	}

	// Convert the map to a slice.
	history := make([]*entities.PuppetHistory, 0, len(historyMap))
	for _, h := range historyMap {
		history = append(history, h)
	}

	return history, nil
}

func (m *mongodbImpl) GetReport(ctx context.Context, id string) (*entities.PuppetReport, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	collection := m.client.Database(mongoDatabase).Collection("reports")

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_report"))
	defer t.ObserveDuration()

	var report entities.PuppetReport
	err := collection.FindOne(ctx, bson.M{"id": bson.M{
		"$eq": id,
		"$ne": "", // This is to ensure that the id is not empty. AKA NOSQL injection.
	}}).Decode(&report)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("error getting report: %w", err)
	}

	return &report, nil
}

func (m *mongodbImpl) GetReports(ctx context.Context, fqdn string) ([]*entities.PuppetReportSummary, error) {
	if fqdn == "" {
		return nil, errors.New("fqdn cannot be empty")
	}

	collection := m.client.Database(mongoDatabase).Collection("reports")

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_reports"))

	cursor, err := collection.Find(ctx, bson.M{"fqdn": bson.M{
		"$eq": fqdn,
		"$ne": "", // This is to ensure that the fqdn is not empty. AKA NOSQL injection.
	}})
	if err != nil {
		return nil, fmt.Errorf("error getting reports: %w", err)
	}

	reports := make([]*entities.PuppetReportSummary, 0)
	if err := cursor.All(ctx, &reports); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error getting reports: %w", err)
	}

	// Rather than defer, we stop the timer here so that we can calculate the time since. This is because we have to
	// iterate over the reports to calculate the time since which would skew the metrics.
	t.ObserveDuration()

	for _, report := range reports {
		report.CalculateTimeSince()
	}

	return reports, nil
}

func (m *mongodbImpl) Ping(ctx context.Context) error {
	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("ping"))
	defer t.ObserveDuration()

	if err := m.client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}
	return nil
}

func (m *mongodbImpl) GetRunsByState(ctx context.Context, states ...summary.State) ([]*entities.PuppetRun, error) {
	collection := m.client.Database(mongoDatabase).Collection("reports")

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_nodes_by_state"))
	defer t.ObserveDuration()

	// Create the filter for the states.
	filter := bson.M{
		"state": bson.M{
			"$in": states,
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error getting nodes: %w", err)
	}

	nodes := make([]*entities.PuppetRun, 0)
	if err := cursor.All(ctx, &nodes); err != nil {
		return nil, fmt.Errorf("error getting nodes: %w", err)
	}

	return nodes, nil
}

func (m *mongodbImpl) GetRuns(ctx context.Context) ([]*entities.PuppetRun, error) {
	collection := m.client.Database(mongoDatabase).Collection("reports")

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_nodes"))
	defer t.ObserveDuration()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error getting nodes: %w", err)
	}

	nodes := make([]*entities.PuppetRun, 0)
	if err := cursor.All(ctx, &nodes); err != nil {
		return nil, fmt.Errorf("error getting nodes: %w", err)
	}

	return nodes, nil
}

func (m *mongodbImpl) SaveRun(ctx context.Context, report *entities.PuppetReport) error {
	collection := m.client.Database(mongoDatabase).Collection("reports")

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("save_run"))
	defer t.ObserveDuration()

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"id": report.ID},
		bson.M{"$set": report},
		opts,
	)
	if err != nil {
		return fmt.Errorf("error saving run: %w", err)
	}
	return nil
}

func NewMongo(ctx context.Context, v *viper.Viper) (Database, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	connectionString := v.GetString("db.conn_str")
	if connectionString != "" {
		slog.Debug("Found MongoDB URI in environment")
	} else {
		// Missing environment variable.
		return nil, fmt.Errorf("missing environment variable: %s", EnvDbConnStr)
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(connectionString).SetServerAPIOptions(serverAPI)
	opts.SetAppName(mongoDatabase)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("connect to MongoDB: %w", err)
	} else if client == nil {
		return nil, errors.New("nil MongoDB client")
	}

	impl := &mongodbImpl{
		client: client,
	}

	return impl, nil
}
