Need to connection

prefix schema: <http://schema.org/>
prefix bds: <http://www.bigdata.com/rdf/search#>
select distinct ?s ?score ?rank
where {
   ?o bds:search "thermal" .  
   ?s rdf:type schema:Dataset .
   ?s schema:description ?o . 
  ?o bds:minRelevance "0.25" .
?o bds:relevance ?score .
?o bds:maxRank "1000" .
?o bds:rank ?rank .
}
ORDER BY DESC(?rank)

prefix schema: <http://schema.org/>
prefix bds: <http://www.bigdata.com/rdf/search#>
select distinct ?pvs  ?s
where {
   ?o bds:search "age" .
   ?s rdf:type schema:Dataset .
   ?s schema:description ?o .
  ?o bds:minRelevance "0.25" .
?o bds:relevance ?score .
?o bds:maxRank "1000" .
?o bds:rank ?rank .
  ?pvs <prov:hadMember> ?s
}
ORDER BY DESC(?rank)



with the prov graph.  For each ?s connection 
to a repo ?name and ?url (minimum).   Will need to 
toss ?score and ?rank since these will make DISTINCT 
pointless.  However, I could COUNT them to give 
a COUNT to each org.  Still use the ORDER by though, 
will it work to order by even with the order by var not 
being in the select (DISTINCT) params?






<a id="top"></a>
## Table of Contents ##
* [Logo](#resource-logo)
* [Repository](#repository)
  * [Services](#repository-services)
    * [Target Properties](#repository-services-target)
    * [Query Inputs](#repository-services-inputs)
* [Dataset](#dataset)
  * [Variables](#dataset-variables)
  * [People](#dataset-people)
  * [Funder](#dataset-funder)
  * [Publisher/Provider](#dataset-publisher_provider)
  * [Query Inputs](#dataset-search-endpoint)
<hr/>


<a id="resource-logo"></a>
### Resource Logo ###

1. When executing this query, we should probably only specify one ?type.
```
PREFIX schema: <http://schema.org/>
SELECT DISTINCT ?graph ?type ?resource ?logo
WHERE {
  VALUES ?resource {
    <http://www.bco-dmo.org/affiliation/191>
  }
  VALUES ?type {
    schema:Organization
    schema:Dataset
  }
  GRAPH ?graph {
    ?resource rdf:type ?type .
    OPTIONAL { ?resource schema:logo [ schema:url ?logo ] }
    
  }
}
ORDER BY ?graph ?type ?resource ?logo
```

<a id="repository"></a>
## Repository ##

<a id="repository-services"></a>
### Repository Services ###

1. ?action_target_type = "@type" means you've got to query down to check for schema:urlTemplate, schema:contentType, and schema:httpMethod. See [Repository Service Target Properties](#repository-services-target)
2. ?action_query_input_type = "@type" means you've got to query down for all the input parameters to a search service. See [Repository Service Query Inputs](#repository-services-inputs)

```
PREFIX schema: <http://schema.org/>
SELECT DISTINCT ?repo ?service ?type ?name ?desc ?service_url ?channel_url 
(IF(isLiteral(?action_target), ?action_target , "@type") as ?action_target_type) 
(IF(isLiteral(?action_query_input),"", "@type") as ?action_query_input_type)
WHERE
{
  VALUES ?type
  {
    "gdx:SearchService"
    "gdx:SubmissionService"
    "gdx:SyndicationService"
  }
  ?repo schema:additionalType "gdx:ResearchRepositoryService" .
  ?repo schema:availableChannel ?channel .
  ?channel schema:providesService ?service .
  OPTIONAL { ?channel schema:serviceUrl ?channel_url }
  ?service schema:additionalType ?type .
  OPTIONAL { ?service schema:name ?name }
  OPTIONAL { ?service schema:description ?desc }
  OPTIONAL { ?service schema:url ?service_url }
  OPTIONAL { 
    ?service schema:potentialAction ?action .
    OPTIONAL { ?action schema:target ?action_target }
    OPTIONAL { 
      ?action a schema:SearchAction .
      ?action schema:query-input ?action_query_input .
    }
  }  
}
ORDER BY ?repo ?type ?service
```

<a id="repository-services-target"></a>
#### Repository Services Target Properties
1. HTTP Method is multi-valued, comma-separated list

```
PREFIX schema: <http://schema.org/>
SELECT DISTINCT ?service ?url_template ?content_type (GROUP_CONCAT(DISTINCT ?http_method ; separator=",") as ?http_method_list)
WHERE
{
  VALUES ?service { <http://lod.bco-dmo.org/sparql> }
  ?service schema:potentialAction [ schema:target ?target ] .
  FILTER EXISTS { ?target rdf:type schema:EntryPoint }
  OPTIONAL { ?target schema:urlTemplate ?url_template }
  OPTIONAL { ?target schema:contentType ?content_type }
  OPTIONAL { ?target schema:httpMethod ?http_method }
 
}
GROUP BY ?service ?url_template ?content_type
ORDER BY ?service
```

<a id="repository-services-inputs"></a>
#### Repository Services Query Inputs ####

```
PREFIX schema: <http://schema.org/>
SELECT DISTINCT ?service ?name ?required ?default ?pattern ?multiple_values ?read_only ?max_value ?min_value ?max_length ?min_length
WHERE
{
  VALUES ?service { <http://lod.bco-dmo.org/sparql> }
  ?service schema:potentialAction ?action .
  ?action a schema:SearchAction .
  ?action schema:query-input ?input .
  ?input a schema:PropertyValueSpecification .
  ?input schema:valueName ?name
  OPTIONAL { ?input schema:valueRequired ?required }
  OPTIONAL { ?input schema:defaultValue ?default }
  OPTIONAL { ?input schema:valuePattern ?pattern }
  OPTIONAL { ?input schema:multipleValues ?multiple_values }
  OPTIONAL { ?input schema:readonlyValue ?read_only }
  OPTIONAL { ?input schema:maxValue ?max_value }
  OPTIONAL { ?input schema:minValue ?min_value }
  OPTIONAL { ?input schema:valueMaxLength ?max_length }
  OPTIONAL { ?input schema:valueMinLength ?min_length }
  
}
ORDER BY ?service ?name
```

<hr/>
<a id="dataset"></a>

## Dataset ##

<a id="dataset-variables"></a>
### Dataset Variables ###

1. I'm not grouping by variable name becuase we can show the need for understanding semantic sameness of variables
```
PREFIX schema: <http://schema.org/>
SELECT DISTINCT ?g ?value ?name ?desc ?url ?units ?var
WHERE
{
  GRAPH ?g {
  VALUES ?dataset
   {
    <https://www.bco-dmo.org/dataset/3300>
    <http://opencoredata.org/id/dataset/bcd15975-680c-47db-a062-ac0bb6e66816>
   }
   ?dataset schema:variableMeasured ?var  .
   OPTIONAL {
    ?var a schema:PropertyValue .
    OPTIONAL { ?var schema:value ?value }
    OPTIONAL { ?var schema:description ?desc }
    OPTIONAL { ?var schema:name ?name }
    OPTIONAL { ?var schema:url ?url }
    OPTIONAL { ?var schema:unitText ?units }
   }
  }
}
ORDER BY lcase(?value) lcase(?name) ?g
```

<a id="dataset-people"></a>
### Dataset People ###

1. Purposefuly, I do not group by ORCiD to demonstrate the NEED for persistent IDs.
2. I do not order the results because there is no guarantee that all person names will get broken out into familyName and givenName, etc. We can improve this query to try to construct name from all fields, for comparison if we have time.

```
PREFIX schema: <http://schema.org/>
SELECT DISTINCT ?g ?person (IF(?role = schema:contributor, "Contributor", IF(?role = schema:creator, "Creator/Author", "Author")) as ?rolename) ?name ?url ?orcid
WHERE
{
  GRAPH ?g {
   VALUES ?dataset
   {
    <https://www.bco-dmo.org/dataset/472032>
    <http://opencoredata.org/id/dataset/bcd15975-680c-47db-a062-ac0bb6e66816>
   }
   VALUES ?role
   {
    schema:author
    schema:creator
    schema:contributor
   }
   { ?dataset ?role ?person }
   OPTIONAL {
    ?person a schema:Person .
    OPTIONAL { ?person schema:name ?name }
    OPTIONAL { ?person schema:url ?url }
    OPTIONAL { 
      ?person schema:identifier ?id .
      ?id schema:propertyID "datacite:orcid" .
      ?id schema:value ?orcid
    }
   }
  }
}
```
<a id="dataset-funder"></a>
### Dataset Funder ###

1. Needs testing

```
PREFIX schema: <http://schema.org/>
PREFIX gdx: <https://geodex.org/voc/>
SELECT DISTINCT ?g ?funder ?legal_name ?name ?url ?award ?award_name ?award_url
WHERE
{
  GRAPH ?g {
   VALUES ?dataset
   {
    <https://www.bco-dmo.org/dataset/472032>
    <http://opencoredata.org/id/dataset/bcd15975-680c-47db-a062-ac0bb6e66816>
   }
   ?dataset schema:funder ?funder
   OPTIONAL {
    ?funder a schema:Organization .
    OPTIONAL { ?funder schema:legalName ?legal_name }
    OPTIONAL { ?funder schema:name ?name }
    OPTIONAL { ?funder schema:url ?url }
    OPTIONAL { 
      ?funder schema:makesOffer ?award .
      ?dataset gdx:fundedBy ?award .
      ?award schema:additionalType "geolink:Award" .
      ?award schema:name ?award_name .
      OPTIONAL { ?award schema:url ?award_url }
    }
   }
  }
}
```

<a id="dataset-publisher_provider"></a>
### Dataset Publisher/Provider ###

```
PREFIX schema: <http://schema.org/>
SELECT DISTINCT ?g ?org ?type (IF(?role = schema:publisher, "Publisher", "Provider") as ?rolename) ?legal_name ?name ?url
WHERE
{
  GRAPH ?g {
   VALUES ?dataset
   {
    <https://www.bco-dmo.org/dataset/472032>
    <http://opencoredata.org/id/dataset/bcd15975-680c-47db-a062-ac0bb6e66816>
   }
   VALUES ?role { schema:publisher  schema:provider }
   VALUES ?type { schema:Organization schema:Person }
   ?dataset ?role ?org .
   ?org a ?type .
   OPTIONAL { ?org schema:legalName ?legal_name }
   OPTIONAL { ?org schema:name ?name }
   OPTIONAL { ?org schema:url ?url }
  }
}
ORDER BY ?role ?legal_name ?name

```

<a id="dataset-search-endpoint"></a>
### Dataset Search Endpoint ###

```
PREFIX schema: <http://schema.org/>
SELECT DISTINCT ?dataset ?action ?action_target ?action_query_input
WHERE
{
  VALUES ?dataset
  {
    <http://some-uri-here>
  }
  ?dataset a schema:Dataset .
  ?dataset schema:potentialAction ?action .
  ?action a schema:SearchAction .
  ?action schema:target ?action_target .
  ?action schema:query-input ?action_query_input .
  OPTIONAL {
    ?action_target schema:urlTemplate ?action_target_url_template .
  }  
}
ORDER BY ?dataset ?action_target
```
Back to [Top](#top)
