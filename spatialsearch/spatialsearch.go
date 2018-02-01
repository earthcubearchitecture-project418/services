package spatialsearch

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/garyburd/redigo/redis"
	"github.com/paulmach/go.geojson"
)

// LocType is there to do a first cut marshalling to just get the type before  next marshalling
type LocType struct {
	Type string `json:"type"`
}

type URLSet []string

// New builds out the services calls..
func New() *restful.WebService {
	service := new(restful.WebService)

	service.
		Path("/api/v1/spatial").
		Doc("Spatial services to P418 holdings").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	service.Route(service.GET("/search/object").To(SpatialCall).
		Doc("get expeditions from a spatial polygon defined by wkt").
		Param(service.QueryParameter("geowithin", "Polygon in WKT format within which to look for features.  Try `POLYGON((-72.2021484375 38.58896696823242,-59.1943359375 38.58896696823242,-59.1943359375 28.11801628757283,-72.2021484375 28.11801628757283,-72.2021484375 38.58896696823242))`").DataType("string")).
		Param(service.QueryParameter("filter", "Filter the URL property in the GeoJSON for the pattern in this parameter if present ").DataType("string")).
		ReturnsError(400, "Unable to handle request", nil).
		Operation("SpatialCall"))

	service.Route(service.GET("/search/resource").To(ResCall).
		Doc("get expeditions from a spatial polygon defined by wkt").
		Param(service.QueryParameter("id", "ID of the resource to locate").DataType("string")).
		ReturnsError(400, "Unable to handle request", nil).
		Operation("ResCall"))

	service.Route(service.POST("/search/resourceset").To(ResSetCall).
		Doc("Call for details on an array of resources from the triplestore (graph)").
		Param(service.BodyParameter("body", "The body containing an array of URIs to obtain parameter values from")).
		Consumes("application/x-www-form-urlencoded").
		ReturnsError(400, "Unable to handle request", nil).
		Operation("ResSetCall"))

	return service
}

func redisDial() (redis.Conn, error) {
	// c, err := redis.Dial("tcp", "tile38:9851")
	c, err := redis.Dial("tcp", "localhost:9851")
	if err != nil {
		log.Printf("Could not connect: %v\n", err)
	}
	return c, err
}

// ResSetCall return the GeoJSON of a set of resources
func ResSetCall(request *restful.Request, response *restful.Response) {
	body, err := request.BodyParameter("body")
	if err != nil {
		log.Printf("Error on body parameter read %v with %s \n", err, body)
	}

	c, err := redisDial()
	defer c.Close()

	ja := []byte(body)
	var jas URLSet
	err = json.Unmarshal(ja, &jas)
	if err != nil {
		log.Println("error with unmarshal..   return http error")
	}

	m := make(map[string]string)
	for k := range jas {
		log.Println(jas[k])
		uri := jas[k]
		uri = strings.Replace(uri, "<", "", -1) // has to be a better way to do this....
		uri = strings.Replace(uri, ">", "", -1)
		log.Printf("Going to get geo for %s \n", uri)
		reply, err := redis.String(c.Do("GET", "p418", uri)) // an early test call just to get everything
		if err != nil {
			fmt.Printf("Error in reply %v \n", err)
		}
		m[uri] = reply
	}

	results, _ := redisStringToGeoJSON(m)
	response.Write([]byte(results))
}

// ResCall return the GeoJSON of a resource
func ResCall(request *restful.Request, response *restful.Response) {
	resid := request.QueryParameter("id")
	log.Println(resid)

	c, err := redisDial()
	defer c.Close()

	reply, err := redis.String(c.Do("GET", "p418", resid)) // an early test call just to get everything
	if err != nil {
		fmt.Printf("Error in reply %v \n", err)
	}

	m := make(map[string]string)
	m[resid] = reply
	results, _ := redisStringToGeoJSON(m)
	response.Write([]byte(results))
}

// SpatialCall calls to tile38 data store
func SpatialCall(request *restful.Request, response *restful.Response) {
	geowithin := request.QueryParameter("geowithin")
	filter := request.QueryParameter("filter")
	// log.Printf("Called with filter: %s and geojson %s \n", filter, geowithin)

	_, err := geojson.UnmarshalFeatureCollection([]byte(geowithin))
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, "Malformed GeoJSON in request, please validate your GeoJSON is a proper FeatureCollection")
	}

	c, err := redisDial()
	defer c.Close()

	var value1 int
	var value2 []interface{}
	// TODO  fix the 50K request limit, put in cursor pattern
	reply, err := redis.Values(c.Do("INTERSECTS", "p418", "LIMIT", "50000", "OBJECT", geowithin))
	// reply, err := redis.Values(c.Do("SCAN", "p418"))  // an early test call just to get everything
	if err != nil {
		fmt.Printf("Error in reply %v \n", err)
	}
	if _, err := redis.Scan(reply, &value1, &value2); err != nil {
		fmt.Printf("Error in scan %v \n", err)
	}

	log.Println(value1) // the point of this logging is what?

	results, _ := redisToGeoJSON(value2, filter)
	response.Write([]byte(results))
}

func redisStringToGeoJSON(m map[string]string) (string, error) {

	fc := geojson.NewFeatureCollection()

	for k, v := range m {
		fmt.Println("k:", k, "v:", v)

		rawGeometryJSON := []byte(v)
		ID := k

		g, err := geojson.UnmarshalGeometry(rawGeometryJSON)
		if err != nil {
			log.Printf("Unmarshal geom error for %s with %s\n", rawGeometryJSON, err)
		}

		switch {
		case g.IsPoint():
			nf := geojson.NewFeature(g)
			nf.SetProperty("URL", ID)
			fc.AddFeature(nf)
		case g.IsPolygon():
			nf := geojson.NewFeature(g)
			nf.SetProperty("URL", ID)
			fc.AddFeature(nf)
		default:
			log.Println(g.Type)
		}

		if g.Type == "Feature" {
			f, err := geojson.UnmarshalFeature(rawGeometryJSON)
			if err != nil {
				log.Printf("Unmarshal feature error for %s with %s\n", ID, err)
			}
			f.SetProperty("URL", ID)
			fc.AddFeature(f)
		}

	}

	rawJSON, err := fc.MarshalJSON()
	if err != nil {
		return "", err
	}

	return string(rawJSON), nil
}

func redisToGeoJSON(results []interface{}, filter string) (string, error) {

	fc := geojson.NewFeatureCollection()

	for _, item := range results {
		valcast := item.([]interface{})
		val0 := fmt.Sprintf("%s", valcast[0])
		val1 := fmt.Sprintf("%s", valcast[1])
		// log.Printf("%s %s \n", val0, val1)

		if strings.Contains(val0, filter) || filter == "" {

			lt := &LocType{}
			err := json.Unmarshal([]byte(val1), lt)
			if err != nil {
				log.Print(err)
				return "", err
			}

			rawGeometryJSON := []byte(val1)

			if lt.Type == "Point" || lt.Type == "Poly" {
				g, err := geojson.UnmarshalGeometry(rawGeometryJSON)
				if err != nil {
					log.Printf("Unmarshal geom error for %s with %s\n", val0, err)
				}

				switch {
				case g.IsPoint():
					nf := geojson.NewFeature(g)
					nf.SetProperty("URL", val0)
					fc.AddFeature(nf)
				case g.IsPolygon():
					nf := geojson.NewFeature(g)
					nf.SetProperty("URL", val0)
					fc.AddFeature(nf)
				default:
					log.Println(g.Type)
				}
			}

			if lt.Type == "Feature" {
				f, err := geojson.UnmarshalFeature(rawGeometryJSON)
				if err != nil {
					log.Printf("Unmarshal feature error for %s with %s\n", val0, err)
				}
				f.SetProperty("URL", val0)
				fc.AddFeature(f)
			}
		}

	}

	rawJSON, err := fc.MarshalJSON()
	if err != nil {
		return "", err
	}

	return string(rawJSON), nil
}
