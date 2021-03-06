package postgis

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/terranodo/tegola"
	"github.com/terranodo/tegola/wkb"
)

func TestForEachFeature(t *testing.T) {
	if os.Getenv("RUN_POSTGIS_TESTS") != "yes" {
		return
	}

	port, err := strconv.ParseInt(os.Getenv("PGPORT"), 10, 64)
	if err != nil {
		t.Fatalf("err parsing PGPORT: %v", err)
	}

	testcases := []struct {
		config       map[string]interface{}
		tile         tegola.Tile
		expectedTags map[string]interface{}
	}{
		{
			config: map[string]interface{}{
				ConfigKeyHost:     os.Getenv("PGHOST"),
				ConfigKeyPort:     port,
				ConfigKeyDB:       os.Getenv("PGDATABASE"),
				ConfigKeyUser:     os.Getenv("PGUSER"),
				ConfigKeyPassword: os.Getenv("PGPASSWORD"),
				ConfigKeyLayers: []map[string]interface{}{
					{
						ConfigKeyLayerName:   "buildings",
						ConfigKeyGeomIDField: "id",
						ConfigKeyGeomField:   "geom",
						ConfigKeySQL:         "SELECT id, height, ST_AsBinary(geom) AS geom FROM hstore_test WHERE geom && !BBOX!",
					},
				},
			},
			tile: tegola.Tile{
				Z: 1,
				X: 1,
				Y: 1,
			},
			expectedTags: map[string]interface{}{
				"height": "10",
			},
		},
	}

	for i, tc := range testcases {
		var err error

		provider, err := NewProvider(tc.config)
		if err != nil {
			t.Errorf("test (%v) failed. Unable to create a new provider. err: %v", i, err)
			return
		}

		p := provider.(Provider)

		//	iterate our configured layers
		for _, tcLayer := range tc.config[ConfigKeyLayers].([]map[string]interface{}) {
			layerName := tcLayer[ConfigKeyLayerName].(string)

			err = p.ForEachFeature(context.Background(), layerName, tc.tile, func(lyr Layer, gid uint64, wgeom wkb.Geometry, ftags map[string]interface{}) error {

				if !reflect.DeepEqual(tc.expectedTags, ftags) {
					t.Errorf("test (%v) failed. expected tags (%+v) does not match output (%+v)", i, tc.expectedTags, ftags)
					return nil
				}

				return nil
			})
			if err != nil {
				t.Errorf("test (%v) failed. err: %v", i, err)
				continue
			}
		}
	}
}
